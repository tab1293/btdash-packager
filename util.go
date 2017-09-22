package main

import (
	"encoding/binary"
	"io"
	"os"
)

func pathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
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
