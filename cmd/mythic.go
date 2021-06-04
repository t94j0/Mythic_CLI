package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(mythicCmd)
	mythicCmd.AddCommand(startMythicCmd)
	mythicCmd.AddCommand(stopMythicCmd)
}

var mythicCmd = &cobra.Command{
	Use:   "mythic",
	Short: "Manage Mythic containers",
}

var startMythicCmd = &cobra.Command{
	Use:   "start",
	Short: "Start Mythic",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.StartMythic(args)
	},
}

var stopMythicCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop Mythic",
	RunE: func(cmd *cobra.Command, args []string) error {
		util.StopMythic(args)
		return nil
	},
}
