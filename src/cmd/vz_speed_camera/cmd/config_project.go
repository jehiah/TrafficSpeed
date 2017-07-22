package cmd

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/spf13/cobra"
	project "lib/vz_project"
)

func configProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config -d project_dir",
		Short: "Configure a project via a HTTP UI",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				log.Fatal(err)
			}
			httpAddress, err := cmd.Flags().GetString("http-address")
			if err != nil {
				log.Fatal(err)
			}

			if projectDir == "" || httpAddress == "" {
				cmd.Usage()
				return
			}
			projectDir, err = filepath.Abs(projectDir)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Printf("opening project dir %s\n", projectDir)
			settingsname := filepath.Join(projectDir, "project.json")
			p, err := project.LoadProject(settingsname)
			if err != nil {
				log.Fatal(err)
			}

			project.ConfigureUI(p, httpAddress)

			// TODO: save p
			// TODO: write and rename
			// f, err := os.Open(settingsname)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// err = json.NewEncoder(f).Encode(p)
			// if err != nil {
			// 	log.Fatal(err)
			// }
			// f.Close()

		},
	}
	cmd.Flags().StringP("dir", "d", "", "project directory configure")
	cmd.Flags().StringP("http-address", "a", ":53001", "http listen address")
	return cmd
}
