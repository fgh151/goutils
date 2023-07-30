package goutils

import (
	"context"
	kl "github.com/go-kit/kit/log"
	sdetcd "github.com/go-kit/kit/sd/etcdv3"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

// healthz is a liveness probe.
func Healthz(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
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

	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}
