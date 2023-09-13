package goutils

import (
	"context"
	"encoding/json"
	"github.com/gin-gonic/gin"
	kl "github.com/go-kit/kit/log"
	sdetcd "github.com/go-kit/kit/sd/etcdv3"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/plugin/prometheus"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// healthz is a liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func HealthzWithDb(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var b bool
		tx := db.Raw("SELECT 1 = 1").Scan(&b)

		if tx.Error != nil {
			c.AbortWithStatus(http.StatusServiceUnavailable)
		}

		c.Writer.WriteHeader(http.StatusOK)
	}
}

// readyz is a readiness probe.
func Readyz(isReady *atomic.Value) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		if isReady == nil || !isReady.Load().(bool) {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func RegisterService(servers []string, prefix string, instance string) (*sdetcd.Registrar, error) {
	key := prefix + instance

	client, err := sdetcd.NewClient(context.Background(), servers, sdetcd.ClientOptions{})
	if err != nil {
		return nil, err
	}

	registrar := sdetcd.NewRegistrar(client, sdetcd.Service{
		Key:   key,
		Value: instance,
	}, GetLogger())

	registrar.Register()

	return registrar, nil
}

func GetLogger() kl.Logger {
	var logger kl.Logger

	logger = kl.NewLogfmtLogger(os.Stderr)
	logger = kl.With(logger, "ts", kl.DefaultTimestampUTC)
	logger = kl.With(logger, "caller", kl.DefaultCaller)

	return logger
}

func InitApp(dsn string) (*gorm.DB, error) {
	_, isCuber := os.LookupEnv("IS_CUBER")

	if false == isCuber {
		err := godotenv.Load()
		if err != nil {
			log.Fatal("Error loading .env file")
		}
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	db.Use(prometheus.New(prometheus.Config{
		DBName:          db.Name(),                   // use `DBName` as metrics label
		RefreshInterval: 15,                          // Refresh metrics interval (default 15 seconds)
		PushAddr:        "prometheus pusher address", // push metrics if `PushAddr` configured
		StartServer:     true,                        // start http server to expose metrics
		HTTPServerPort:  8080,                        // configure http server port, default port 8080 (if you have configured multiple instances, only the first `HTTPServerPort` will be used to start server)
		MetricsCollector: []prometheus.MetricsCollector{
			&prometheus.Postgres{
				VariableNames: []string{"Threads_running"},
			},
		}, // user defined metrics
	}))

	return db, err
}

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
