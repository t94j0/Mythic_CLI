package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(logsCmd)
}

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Mythic logs",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		util.GetLogs(args[0])
	},
}
