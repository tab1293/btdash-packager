package main

import (
	"flag"
	"fmt"
	"os"
	"path"
)

var Args struct {
	Output         string
	ForceTranscode bool
	JSONSideCar    bool
}

func init() {
	flag.BoolVar(&Args.ForceTranscode, "force-transcode", false, "Should the input file be transmuxed regardless of if it has to")
	flag.BoolVar(&Args.JSONSideCar, "json", false, "Output JSON formatted manifest")
	flag.StringVar(&Args.Output, "output", "./", "Output directory for torrent and video files")
}

func main() {
	flag.Parse()

	if len(os.Args) < 2 {
		fmt.Printf("No input specified\n")
		os.Exit(1)
	}

	input := os.Args[1]
	if input == "" {
		fmt.Printf("No input specified\n")
		os.Exit(1)
	}

	ffmpegCmd, err := FFmpegCmd(input, Args.ForceTranscode, Args.Output)
	if err != nil {
		fmt.Printf("FFmpeg command error: %s\n", err)
		os.Exit(1)
	}

	if ffmpegCmd != nil {
		ffmpegCmd.Run()
	}

	outputVideoFile := path.Join(Args.Output, "out.mp4")
	outputTorrentFile := path.Join(Args.Output, "out.torrent")

	segments := GetSegments(outputVideoFile)
	CreateTorrentFile(outputVideoFile, segments, outputTorrentFile)

	if Args.JSONSideCar {
		outputJsonFile := path.Join(Args.Output, "out.json")
		WriteSegmentsToFile(segments, outputJsonFile)
	}
}
