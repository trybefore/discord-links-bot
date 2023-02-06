package health

import (
	"context"
	"log"
	"net/http"

	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/config"
)

func Run(context.Context) error {
	http.HandleFunc("/health_check", Check)

	if err := http.ListenAndServe(":"+viper.GetString(config.HealthCheckPort), nil); err != nil {
		log.Fatal(err)
	}

	return nil
}

func Check(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
