package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	project "lib/vz_project"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new -f video.m4a -d project_dir",
		Short: "Start a new project",
		Run: func(cmd *cobra.Command, args []string) {
			targetDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				log.Fatal(err)
			}
			filename, err := cmd.Flags().GetString("file")
			if err != nil {
				log.Fatal(err)
			}

			if filename == "" || targetDir == "" {
				cmd.Usage()
				return
			}
			targetDir, err = filepath.Abs(targetDir)
			if err != nil {
				log.Fatal(err)
			}
			err = os.MkdirAll(targetDir, 0777)
			if err != nil {
				log.Fatal(err)
			}

			filename, err = filepath.Abs(filename)
			if err != nil {
				log.Fatal(err)
			}
			relfilename, err := filepath.Rel(targetDir, filename)
			if err != nil {
				log.Fatal(err)
			}

			iterator, err := project.NewIterator(filename)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("detected video at resolution %s\n", iterator.VideoResolution())
			iterator.Close()

			p := project.NewProject(filename)
			p.Filename = relfilename

			settingsname := filepath.Join(targetDir, "project.json")
			fmt.Printf("creating %s\n", settingsname)
			f, err := os.Create(settingsname)
			if err != nil {
				log.Fatal(err)
			}

			err = json.NewEncoder(f).Encode(p)
			if err != nil {
				log.Fatal(err)
			}
			f.Close()
		},
	}
	cmd.Flags().StringP("file", "f", "", "video filename")
	cobra.MarkFlagFilename(cmd.Flags(), "file", "m4a")
	cmd.Flags().StringP("dir", "d", "", "directory to output to")
	return cmd
}
