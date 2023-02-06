package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/trybefore/linksbot/internal/bot"
	"github.com/trybefore/linksbot/internal/config"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use: "run",
	RunE: func(cmd *cobra.Command, args []string) error {
		return bot.Run(cmd.Context())
	},
}

func init() {
	runCmd.PersistentFlags().String(config.BotToken, "", "discord bot token, with or without preceding Bot")
	runCmd.PersistentFlags().String(config.HealthCheckPort, ":8800", "the port to listen to for health checks")

	if err := viper.BindPFlags(runCmd.Flags()); err != nil {
		log.Fatal(err)
	}

	rootCmd.AddCommand(runCmd)
}
