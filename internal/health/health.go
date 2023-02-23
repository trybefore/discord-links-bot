package health

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/config"
)

func Run(ctx context.Context) error {
	http.HandleFunc("/health_check", handleHealthCheck)

	server := &http.Server{
		Addr:    ":" + viper.GetString(config.HealthCheckPort),
		Handler: http.DefaultServeMux,
	}

	go func() {
		log.Printf("health check starting at %s", server.Addr)
		if err := server.ListenAndServe(); err != nil {
			log.Printf("listenAndServe: %v", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	log.Printf("shutting down health server...")

	return server.Shutdown(ctx)
}

func handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
