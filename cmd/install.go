package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

var force bool

func init() {
	rootCmd.AddCommand(installCmd)
	installCmd.AddCommand(githubInstallCmd)
	installCmd.AddCommand(folderInstallCmd)
	installCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "forces the removal of the currently installed version and overwrites with the new, otherwise will prompt you")
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install external container",
}

var githubInstallCmd = &cobra.Command{
	Use:   "github <url> [branch name]",
	Short: "Install container from GitHub",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.InstallAgent(args[0], force)
	},
}

var folderInstallCmd = &cobra.Command{
	Use:   "folder <path to folder>",
	Short: "Install container from folder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.InstallFolder(args[0], force)
	},
}
