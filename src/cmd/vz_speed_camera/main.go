package main

import (
	"fmt"
	"log"
	"os"

	"cmd/vz_speed_camera/cmd"
)

// #cgo LDFLAGS="-L/usr/local/Cellar/ffmpeg/3.3/lib"
// #cgo CGO_CFLAGS="-I/usr/local/Cellar/ffmpeg/3.3/include"

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds | log.Llongfile)
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
