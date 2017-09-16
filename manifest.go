package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/big"
	"os"
	"time"
)

func scaleTimeSlow(ot int64, timescale uint32) time.Duration {
	t := big.NewInt(ot)
	t.Mul(t, big.NewInt(int64(time.Second)))
	t.Div(t, big.NewInt(int64(timescale)))
	return time.Duration(t.Int64())
}

func ScaleTime(ot int64, timescale uint32) time.Duration {
	if ot < math.MaxInt64/int64(time.Second) {
		return time.Duration((ot * int64(time.Second)) / int64(timescale))
	} else {
		return scaleTimeSlow(ot, timescale)
	}
}

type Segment struct {
	Start     int64
	End       int64
	StartTime int64
	EndTime   int64
}

func readByte(r io.Reader) uint32 {
	b := []byte{0}
	io.ReadFull(r, b)
	return uint32(b[0])
}

func read16(r io.Reader) uint32 {
	var v uint16
	binary.Read(r, binary.BigEndian, &v)
	return uint32(v)
}

func read24(r io.Reader) uint32 {
	b := []byte{0, 0, 0}
	io.ReadFull(r, b)
	return (uint32(b[0]) << 16) | (uint32(b[1]) << 8) | uint32(b[2])
}

func read32(r io.Reader) uint32 {
	var v uint32
	binary.Read(r, binary.BigEndian, &v)
	return v
}

func read64(r io.Reader) uint64 {
	var v uint64
	binary.Read(r, binary.BigEndian, &v)
	return v
}

func GetSegments(filePath string) []Segment {
	f, err := os.Open(filePath)
	if err != nil {
		panic(err)
	}

	fi, _ := os.Stat(filePath)
	fileSize := fi.Size()

	var segments []Segment
	segmentCount := 0
	runningDuration, _ := time.ParseDuration("0s")
	firstSidx := true
	boxHeader := make([]byte, 8)
	var pos int64

	for {
		if fileSize-pos < 8 {
			break
		}
		n, err := f.Read(boxHeader)
		if n < 8 {
			panic(io.ErrUnexpectedEOF)
		}
		if err != nil {
			panic(err)
		}

		size := int64(binary.BigEndian.Uint32(boxHeader[0:4]))
		tag := string(boxHeader[4:8])

		if size < 8 {
			panic(fmt.Errorf("invalid box size"))
		}

		if tag == "sidx" && firstSidx {
			seg := Segment{}
			seg.Start = pos
			if segmentCount > 0 {
				segments[segmentCount-1].End = pos - 1
			}
			segmentCount++

			b := make([]byte, size-8)
			_, err = io.ReadFull(f, b)
			if err != nil {
				panic(err)
			}
			buf := bytes.NewBuffer(b)

			version := readByte(buf) // version
			read24(buf)              // flags
			read32(buf)              // referenceId
			timescale := read32(buf)

			if version == 0 {
				read32(buf) // earliest presentation time
				read32(buf) // first offset
			} else {
				read64(buf) // earliest presentation time
				read64(buf) // first offset
			}
			read16(buf) // reserved
			read16(buf) // count int

			read32(buf)
			seg.StartTime = runningDuration.Nanoseconds() / 1e6
			duration := int64(read32(buf))
			t := ScaleTime(duration, timescale)
			runningDuration += t
			seg.EndTime = runningDuration.Nanoseconds() / 1e6

			segments = append(segments, seg)
		}

		pos += size

		if tag == "sidx" {
			firstSidx = !firstSidx
		}

		if pos >= fileSize {
			break
		}
		_, err = f.Seek(pos, 0)
		if err != nil {
			break
		}
	}

	segments[segmentCount-1].End = fileSize
	return segments
}

func WriteSegmentsToFile(segments []Segment, outputFile string) error {
	f, err := os.Create(outputFile)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	return enc.Encode(segments)
}
