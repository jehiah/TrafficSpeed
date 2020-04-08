package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(extractFramesCmd())
	RootCmd.AddCommand(newProjectCmd())
	RootCmd.AddCommand(configProjectCmd())
}

var RootCmd = &cobra.Command{
	Use:   "speed_camera",
	Short: "Speed Camera measure traffic speed",
	Long:  "Speed Camera is a tool to measure vehicle speed from video footage.",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Usage()
			os.Exit(1)
		}
	},
}
