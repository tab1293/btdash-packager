package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	bencode "github.com/jackpal/bencode-go"
	"io"
	"os"
	"path"
	"runtime"
)

const MinimumPieceLength = 16 * 1024
const TargetPieceCountLog2 = 10
const TargetPieceCountMin = 1 << TargetPieceCountLog2

// Target piece count should be < TargetPieceCountMax
const TargetPieceCountMax = TargetPieceCountMin << 1

// Choose a good piecelength.
func choosePieceLength(totalLength int64) (pieceLength int64) {
	// Must be a power of 2.
	// Must be a multiple of 16KB
	// Prefer to provide around 1024..2048 pieces.
	pieceLength = MinimumPieceLength
	pieces := totalLength / pieceLength
	for pieces >= TargetPieceCountMax {
		pieceLength <<= 1
		pieces >>= 1
	}
	return
}

type chunk struct {
	i    int64
	data []byte
}

// computeSums reads the file content and computes the SHA1 hash for each
// piece. Spawns parallel goroutines to compute the hashes, since each
// computation takes ~30ms.
func computeSums(filePath string, totalLength int64, pieceLength int64) (sums []byte, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	// Calculate the SHA1 hash for each piece in parallel goroutines.
	hashes := make(chan chunk)
	results := make(chan chunk, 3)
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		go hashPiece(hashes, results)
	}

	// Read file content and send to "pieces", keeping order.
	numPieces := (totalLength + pieceLength - 1) / pieceLength
	go func() {
		for i := int64(0); i < numPieces; i++ {
			piece := make([]byte, pieceLength, pieceLength)
			if i == numPieces-1 {
				piece = piece[0 : totalLength-i*pieceLength]
			}
			// Ignore errors.
			f.ReadAt(piece, i*pieceLength)
			hashes <- chunk{i: i, data: piece}
		}
		close(hashes)
	}()

	// Merge back the results.
	sums = make([]byte, sha1.Size*numPieces)
	for i := int64(0); i < numPieces; i++ {
		h := <-results
		copy(sums[h.i*sha1.Size:], h.data)
	}
	return
}

func hashPiece(h chan chunk, result chan chunk) {
	hasher := sha1.New()
	for piece := range h {
		hasher.Reset()
		_, err := hasher.Write(piece.data)
		if err != nil {
			result <- chunk{piece.i, nil}
		} else {
			result <- chunk{piece.i, hasher.Sum(nil)}
		}
	}
}

type FileDict struct {
	Length int64
	Path   []string
	Md5sum string
}

type InfoDict struct {
	PieceLength int64 `bencode:"piece length"`
	Pieces      string
	Private     int64
	Name        string
	// Single File Mode
	Length int64
	Md5sum string
	// Multiple File mode
	Files []FileDict
}

func (i *InfoDict) toMap() (m map[string]interface{}) {
	id := map[string]interface{}{}
	// InfoDict
	if i.PieceLength != 0 {
		id["piece length"] = i.PieceLength
	}
	if i.Pieces != "" {
		id["pieces"] = i.Pieces
	}
	if i.Private != 0 {
		id["private"] = i.Private
	}
	if i.Name != "" {
		id["name"] = i.Name
	}
	if i.Length != 0 {
		id["length"] = i.Length
	}
	if i.Md5sum != "" {
		id["md5sum"] = i.Md5sum
	}
	if len(i.Files) > 0 {
		var fi []map[string]interface{}
		for ii := range i.Files {
			f := &i.Files[ii]
			fd := map[string]interface{}{}
			if f.Length > 0 {
				fd["length"] = f.Length
			}
			if len(f.Path) > 0 {
				fd["path"] = f.Path
			}
			if f.Md5sum != "" {
				fd["md5sum"] = f.Md5sum
			}
			if len(fd) > 0 {
				fi = append(fi, fd)
			}
		}
		if len(fi) > 0 {
			id["files"] = fi
		}
	}
	if len(id) > 0 {
		m = id
	}
	return
}

type MetaInfo struct {
	Info         InfoDict
	InfoHash     string
	Announce     string
	AnnounceList [][]string `bencode:"announce-list"`
	CreationDate string     `bencode:"creation date"`
	Comment      string
	CreatedBy    string `bencode:"created by"`
	Encoding     string

	Manifest Manifest
}

func (m *MetaInfo) Bencode(w io.Writer) (err error) {
	var mi map[string]interface{} = map[string]interface{}{}
	id := m.Info.toMap()
	if len(id) > 0 {
		mi["info"] = id
	}
	// Do not encode InfoHash. Clients are supposed to calculate it themselves.
	if m.Announce != "" {
		mi["announce"] = m.Announce
	}
	if len(m.AnnounceList) > 0 {
		mi["announce-list"] = m.AnnounceList
	}
	if m.CreationDate != "" {
		mi["creation date"] = m.CreationDate
	}
	if m.Comment != "" {
		mi["comment"] = m.Comment
	}
	if m.CreatedBy != "" {
		mi["created by"] = m.CreatedBy
	}
	if m.Encoding != "" {
		mi["encoding"] = m.Encoding
	}

	mi["manifest"] = m.Manifest.toMap()

	bencode.Marshal(w, mi)
	return
}

func (m *MetaInfo) UpdateInfoHash() (err error) {
	var b bytes.Buffer
	infoMap := m.Info.toMap()
	if len(infoMap) > 0 {
		err = bencode.Marshal(&b, infoMap)
		if err != nil {
			return
		}
	}
	hash := sha1.New()
	hash.Write(b.Bytes())

	m.InfoHash = string(hash.Sum(nil))
	return
}

func CreateTorrentFile(inputFile string, manifest Manifest, outputTorrentFile string) {
	fi, err := os.Stat(inputFile)
	if err != nil {
		os.Exit(1)
	}

	i := InfoDict{}
	_, name := path.Split(inputFile)
	i.Name = name
	i.Length = fi.Size()
	i.PieceLength = choosePieceLength(i.Length)

	pieceSum, err := computeSums(inputFile, i.Length, i.PieceLength)
	if err != nil {
		fmt.Printf("Error computing piece sum: %s", inputFile)
		os.Exit(1)
	}

	i.Pieces = string(pieceSum)

	m := MetaInfo{}
	m.Info = i
	m.UpdateInfoHash()
	m.Announce = "http://tracker.vanitycore.co:6969/announce"
	m.Manifest = manifest
	// m.Segments = segments

	f, err := os.Create(outputTorrentFile)
	defer f.Close()

	err = m.Bencode(f)
	if err != nil {
		fmt.Printf("Error bencoding torrent file %s", outputTorrentFile)
		os.Exit(1)
	}

}

func decode() {
	f, err := os.Open("test.torrent")
	if err != nil {
		panic(err)
	}

	m := MetaInfo{}
	bencode.Unmarshal(f, &m)

	fmt.Printf("%v\n", m)
}
