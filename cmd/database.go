package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(databaseCmd)
	databaseCmd.AddCommand(databaseResetCmd)
}

var databaseCmd = &cobra.Command{
	Use:   "database",
	Short: "Configure database",
}

var databaseResetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		util.DatabaseReset()
	},
}
