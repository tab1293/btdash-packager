package main

import (
	"fmt"
	// "encoding/json"
	"flag"
	"os"
	"os/exec"
	"path"
)

var Args struct {
	Input    string
	Output   string
	Transmux bool
}

func init() {
	flag.StringVar(&Args.Input, "input", "", "Input file")
	flag.BoolVar(&Args.Transmux, "transmux", false, "Should the input file be transmuxed")
	flag.StringVar(&Args.Output, "output", "", "Output directory for torrent")
}

func main() {
	flag.Parse()

	if Args.Input == "" {
		fmt.Printf("No input specified\n")
		os.Exit(1)
	}

	if Args.Output == "" {
		Args.Output = "/tmp/btdash"
	}

	exists, err := pathExists(Args.Output)
	if !exists || err != nil {
		os.Mkdir(Args.Output, 0755)
	}

	var ffmpegCmd *exec.Cmd
	if Args.Transmux {
		ffmpegCmd = transmuxCmd(Args.Input, path.Join(Args.Output, "out.mp4"))
	}

	if ffmpegCmd != nil {
		ffmpegCmd.Run()
	}

	CreateTorrentFile(path.Join(Args.Output, "out.mp4"), path.Join(Args.Output, "out.torrent"))

	// segments := generateManifest(path.Join(Args.Output, "out.mp4"))

	// b, _ := json.MarshalIndent(segments, "", "    ")
	// fmt.Printf("%s\n", string(b))
}
