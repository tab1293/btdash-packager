package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strconv"
)

type FFprobeStream struct {
	Type       string `json:"codec_type"`            // audio or video
	Index      int    `json:"index"`                 // stream index
	CodecName  string `json:"codec_name"`            // h264 or aac
	Profile    string `json:"profile,omitempty"`     // High
	Width      int    `json:"width,omitempty"`       // 1280
	Height     int    `json:"height,omitempty"`      // 720
	SampleRate string `json:"sample_rate,omitempty"` // 48000
}

type FFprobeFormat struct {
	Filename string `json:"filename"`
	Duration string `json:"duration"`
	Size     string `json:"size"`
	Bitrate  string `json:"bit_rate"`
}

type FFprobe struct {
	Streams []FFprobeStream `json:"streams"`
	Format  *FFprobeFormat  `json:"format"`
}

type AVFileInfo struct {
	Video    FFprobeStream
	Audio    FFprobeStream
	Duration float64
	Size     int64
	Bitrate  int64
}

func GetAVFileInfo(filePath string) (*AVFileInfo, error) {
	var arg []string
	arg = append(arg, "-v", "quiet")
	arg = append(arg, "-print_format", "json")
	arg = append(arg, "-show_format")
	arg = append(arg, "-show_streams")
	arg = append(arg, filePath)
	fmt.Printf("%s\n", arg)

	out, err := exec.Command("ffprobe", arg...).Output()
	if err != nil {
		return nil, err
	}

	p := &FFprobe{}
	err = json.Unmarshal(out, p)
	if err != nil {
		return nil, err
	}

	fi := &AVFileInfo{}
	for i := range p.Streams {
		switch p.Streams[i].Type {
		case "audio":
			fi.Audio = p.Streams[i]
		case "video":
			fi.Video = p.Streams[i]
		}
	}

	fi.Duration, err = strconv.ParseFloat(p.Format.Duration, 64)
	if err != nil {
		return nil, err
	}

	fi.Size, err = strconv.ParseInt(p.Format.Size, 10, 64)
	if err != nil {
		return nil, err
	}

	fi.Bitrate, err = strconv.ParseInt(p.Format.Bitrate, 10, 64)
	if err != nil {
		return nil, err
	}

	return fi, nil
}

func FFmpegCmd(filePath string, forceTranscode bool, outputPath string) (*exec.Cmd, error) {
	fi, err := GetAVFileInfo(filePath)
	if err != nil {
		return nil, err
	}

	videoCodecName := "h264"
	if fi.Video.CodecName == videoCodecName {
		videoCodecName = "copy"
	}

	audiCodecName := "libfdk_aac"
	if fi.Audio.CodecName == audiCodecName {
		audiCodecName = "copy"
	}

	var arg []string
	arg = append(arg, "-y")
	arg = append(arg, "-i", filePath)
	arg = append(arg, "-c:a", audiCodecName)
	arg = append(arg, "-c:v", videoCodecName)
	arg = append(arg, "-f", "mp4")
	arg = append(arg, "-movflags", "+skip_trailer+dash")
	arg = append(arg, "-frag_duration", "6000000")
	arg = append(arg, path.Join(outputPath, "out.mp4"))

	fmt.Printf("%s\n", arg)
	cmd := exec.Command("ffmpeg", arg...)
	cmd.Stderr = os.Stdout
	return cmd, nil
}
