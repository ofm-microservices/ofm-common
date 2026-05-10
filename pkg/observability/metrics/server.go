package metrics

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ofm-microservices/ofm-common/pkg/logging"
)

// StartServer starts the metrics HTTP listener and blocks until the server
// exits.
func StartServer(ctx context.Context, cfg Config, meter Meter, log logging.Logger) error {
	if !cfg.Enabled {
		return nil
	}
	if meter == nil {
		return ErrNilMeter
	}
	if cfg.Path == "" {
		cfg.Path = "/metrics"
	}

	mux := http.NewServeMux()
	mux.Handle(cfg.Path, meter.Handler())
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	if log != nil {
		log.Info("starting metrics server", logging.String("addr", srv.Addr), logging.String("path", cfg.Path))
	}
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
