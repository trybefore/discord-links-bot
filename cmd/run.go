package cmd

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/bot"
	"github.com/trybefore/linksbot/internal/config"
	"github.com/trybefore/linksbot/internal/health"
	"golang.org/x/sync/errgroup"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := run(cmd.Context())

		if !errors.Is(err, http.ErrServerClosed) && !errors.Is(err, context.Canceled) {
			return err
		}

		return nil
	},
}

func run(ctx context.Context) error {
	errs, ctx := errgroup.WithContext(ctx)

	errs.SetLimit(2)

	errs.Go(func() error {
		return bot.Run(ctx)
	})
	if !viper.GetBool(config.DisableHealthCheck) {
		errs.Go(func() error {
			return health.Run(ctx)
		})
	}

	return errs.Wait()
}

func init() {
	runCmd.PersistentFlags().String(config.BotToken, "", "discord bot token, with or without preceding Bot")
	runCmd.PersistentFlags().String(config.HealthCheckPort, ":8800", "the port to listen to for health checks")
	runCmd.PersistentFlags().Bool(config.DisableHealthCheck, false, "don't start the http server with health endpoint")

	if err := viper.BindPFlags(runCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(runCmd)
}
