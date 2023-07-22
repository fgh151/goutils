package goutils

import (
	"context"
	kl "github.com/go-kit/kit/log"
	sdetcd "github.com/go-kit/kit/sd/etcdv3"
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

func RegisterService(prefix string, instance string) (*sdetcd.Registrar, error) {
	var (
		etcdServer = os.Getenv("ETCD_ADDR")
		key        = prefix + instance
	)

	client, err := sdetcd.NewClient(context.Background(), []string{etcdServer}, sdetcd.ClientOptions{})
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
