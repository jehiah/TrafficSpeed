package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of speed_camera",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("speed_camera v0.1 -- HEAD")
	},
}
