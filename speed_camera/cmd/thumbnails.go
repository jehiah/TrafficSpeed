package cmd

import (
	"fmt"
	"image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/jehiah/TrafficSpeed/internal/project"
	"github.com/spf13/cobra"
)

func extractFramesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "thumbnails -f video.m4a",
		Short: "Extract png images of n frames",
		Run: func(cmd *cobra.Command, args []string) {
			extractCount, err := cmd.Flags().GetInt("number")
			if err != nil {
				log.Fatal(err)
			}
			extractDir, err := cmd.Flags().GetString("dir")
			if err != nil {
				log.Fatal(err)
			}
			extractSeek, err := cmd.Flags().GetDuration("seek")
			if err != nil {
				log.Fatal(err)
			}
			filename, err := cmd.Flags().GetString("file")
			if err != nil {
				log.Fatal(err)
			}

			if extractCount <= 0 || filename == "" {
				cmd.Usage()
				return
			}

			iterator, err := project.NewIterator(filename)
			if err != nil {
				log.Fatal(err)
			}
			defer iterator.Close()
			fmt.Printf("extracting thumbnails at resolution %s\n", iterator.VideoResolution())

			if extractSeek > 0 {
				err = iterator.Seek(extractSeek)
				if err != nil {
					log.Fatal(err)
				}
			}

			for iterator.Next() {
				if extractCount > 0 && iterator.Frame() > extractCount {
					break
				}
				img := iterator.Image()
				if img == nil {
					continue
				}
				pngname := filepath.Join(extractDir, fmt.Sprintf("%06d.png", iterator.Frame()))
				fmt.Printf("creating %s\n", pngname)
				f, err := os.Create(pngname)
				if err != nil {
					log.Fatal(err)
				}
				png.Encode(f, img)
				f.Close()
			}

		},
	}
	cmd.Flags().StringP("file", "f", "", "video filename")
	cobra.MarkFlagFilename(cmd.Flags(), "file", "m4a")
	cmd.Flags().Int("number", 5, "number of frames to extract (-1 for all)")
	cmd.Flags().String("dir", "", "directory to output images")
	cmd.Flags().Duration("seek", 0, "seek before extracting frames")
	return cmd
}
