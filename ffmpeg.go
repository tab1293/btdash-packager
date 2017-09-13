package main

import (
	"os"
	"os/exec"
)

func transmuxCmd(filePath string, outputPath string) *exec.Cmd {
	var arg []string
	arg = append(arg, "-y")
	arg = append(arg, "-i", filePath)
	arg = append(arg, "-c", "copy")
	arg = append(arg, "-f", "mp4")
	arg = append(arg, "-movflags", "+skip_trailer+dash")
	arg = append(arg, "-frag_duration", "6000000")
	arg = append(arg, outputPath)

	cmd := exec.Command("ffmpeg", arg...)
	cmd.Stderr = os.Stdout
	return cmd
}
