package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "vz_speed_camera",
	Short: "VZ Speed Camera measure traffic speed",
	Long: "Vision Zero Speed Camera is a tool to measure vehicle speed from video footage.",
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
