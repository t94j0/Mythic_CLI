package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(rabbitMqCmd)
	rabbitMqCmd.AddCommand(resetRabbitMqCmd)
}

var rabbitMqCmd = &cobra.Command{
	Use:   "rabbitmq",
	Short: "Configure RabbitMQ",
}

var resetRabbitMqCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset RabbitMQ",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		util.RabbitmqReset()
	},
}
