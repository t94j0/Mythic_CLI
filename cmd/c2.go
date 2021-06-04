package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(c2Cmd)
	c2Cmd.AddCommand(startC2Cmd)
	c2Cmd.AddCommand(stopC2Cmd)
	c2Cmd.AddCommand(addC2Cmd)
	c2Cmd.AddCommand(removeC2Cmd)
	c2Cmd.AddCommand(listC2Cmd)
}

var c2Cmd = &cobra.Command{
	Use:   "c2",
	Short: "Manage C2 containers",
}

var startC2Cmd = &cobra.Command{
	Use:   "start [c2 profile...]",
	Short: "Start C2 container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.StartMythic(args)
	},
}

var stopC2Cmd = &cobra.Command{
	Use:   "stop [c2 profile...]",
	Short: "Stop C2 container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.StopC2(args)
	},
}

var addC2Cmd = &cobra.Command{
	Use:   "add [c2 profile...]",
	Short: "Add C2 container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.AddRemoveDockerComposeEntries("add", "c2", args)
	},
}

var removeC2Cmd = &cobra.Command{
	Use:   "remove [c2 profile...]",
	Short: "Remove C2 container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.AddRemoveDockerComposeEntries("remove", "c2", args)
	},
}

var listC2Cmd = &cobra.Command{
	Use:   "list [c2 profile...]",
	Short: "List C2 container",
	Run: func(cmd *cobra.Command, args []string) {
		util.ListGroupEntries("c2")
	},
}
