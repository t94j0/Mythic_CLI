package cmd

import (
	"github.com/spf13/cobra"
	"github.com/t94j0/Mythic_CLI/util"
)

func init() {
	rootCmd.AddCommand(payloadCmd)
	payloadCmd.AddCommand(startPayloadCmd)
	payloadCmd.AddCommand(stopPayloadCmd)
	payloadCmd.AddCommand(addPayloadCmd)
	payloadCmd.AddCommand(removePayloadCmd)
	payloadCmd.AddCommand(listPayloadCmd)
}

var payloadCmd = &cobra.Command{
	Use:   "payload",
	Short: "Manage payload containers",
}

var startPayloadCmd = &cobra.Command{
	Use:   "start [payload name]",
	Short: "Start payload container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.StartPayload(args)
	},
}

var stopPayloadCmd = &cobra.Command{
	Use:   "stop [payload name]",
	Short: "Stop payload container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.StopPayload(args)
	},
}

var addPayloadCmd = &cobra.Command{
	Use:   "add [payload name]",
	Short: "Add payload container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.AddRemoveDockerComposeEntries("add", "payload", args)
	},
}

var removePayloadCmd = &cobra.Command{
	Use:   "remove [payload name]",
	Short: "Remove payload container",
	RunE: func(cmd *cobra.Command, args []string) error {
		return util.AddRemoveDockerComposeEntries("remove", "payload", args)
	},
}

var listPayloadCmd = &cobra.Command{
	Use:   "list [payload name]",
	Short: "List payload container",
	Run: func(cmd *cobra.Command, args []string) {
		util.ListGroupEntries("payload")
	},
}
