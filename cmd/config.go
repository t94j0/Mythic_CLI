package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(getConfigCmd)
	configCmd.AddCommand(setConfigCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configure Mythic",
}

var getConfigCmd = &cobra.Command{
	Use:   "get [name...]",
	Short: "Get Mythic configuration",
	Run: func(cmd *cobra.Command, args []string) {
		util.GetEnv(args)
	},
}

var setConfigCmd = &cobra.Command{
	Use:   "set <name> <value>",
	Short: "Set Mythic configuration",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		util.SetEnv(args[0], args[1])
	},
}
