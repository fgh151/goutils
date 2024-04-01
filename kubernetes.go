package sdk

import (
	"github.com/gin-gonic/gin"
	kl "github.com/go-kit/kit/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
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

func GetLogger() kl.Logger {
	var logger kl.Logger

	logger = kl.NewLogfmtLogger(os.Stderr)
	logger = kl.With(logger, "ts", kl.DefaultTimestampUTC)
	logger = kl.With(logger, "caller", kl.DefaultCaller)

	return logger
}

func AppendMetrics(r *gin.Engine) {
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
}
