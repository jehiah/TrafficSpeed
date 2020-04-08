package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/jehiah/TrafficSpeed/internal/project"
	"github.com/spf13/cobra"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "new -f video.m4a -d project_dir",
		Short: "Start a new project",
		Run: func(cmd *cobra.Command, args []string) {
			projectDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				log.Fatal(err)
			}
			filename, err := cmd.Flags().GetString("file")
			if err != nil {
				log.Fatal(err)
			}

			if filename == "" || projectDir == "" {
				cmd.Usage()
				return
			}
			projectDir, err = filepath.Abs(projectDir)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("making project dir %s\n", projectDir)
			err = os.MkdirAll(projectDir, 0777)
			if err != nil {
				log.Fatal(err)
			}

			filename, err = filepath.Abs(filename)
			if err != nil {
				log.Fatal(err)
			}
			relfilename, err := filepath.Rel(projectDir, filename)
			if err != nil {
				log.Fatal(err)
			}

			iterator, err := project.NewIterator(filename)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("detected video at resolution %s\n", iterator.VideoResolution())
			defer iterator.Close()

			p := project.NewProject(filename, iterator)
			p.Dir = projectDir
			p.Filename = relfilename

			fmt.Printf("creating project.json\n")
			settingsname := filepath.Join(projectDir, "project.json")
			f, err := os.Create(settingsname)
			if err != nil {
				log.Fatal(err)
			}
			err = json.NewEncoder(f).Encode(p)
			if err != nil {
				log.Fatal(err)
			}
			f.Close()

			fmt.Printf("extracting first frame as base.png\n")
			err = p.SaveImage(iterator.Image(), "base.png")
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	cmd.Flags().StringP("file", "f", "", "video filename")
	cobra.MarkFlagFilename(cmd.Flags(), "file", "m4a")
	cmd.Flags().StringP("dir", "d", "", "directory to output to")
	return cmd
}
