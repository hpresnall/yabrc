package index

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	gopath "path"
	"time"

	humanize "github.com/dustin/go-humanize"
	"golang.org/x/text/unicode/norm"
)

// Entry represents the data for a single file in the Index.
type Entry struct {
	path    string
	lastMod time.Time // file modification time
	size    int64     // file size in bytes
	hash    string    // hash of file contents
}

// internal use only; Entries should only be created by Index
func buildEntry(path string, info os.FileInfo) (Entry, error) {
	var e Entry

	if path == "" {
		return e, errors.New("path cannot be empty")
	}

	if info == nil {
		return e, errors.New("info cannot be nil")
	}

	base := gopath.Base(path)
	if base != norm.NFC.String(info.Name()) {
		return e, fmt.Errorf("path '%s' does not match FileInfo.Name() '%s'", path, info.Name())
	}

	// read all of the file into sha256; use the actual file name, not the normalized path
	file, err := indexFs.Open(gopath.Join(gopath.Dir(path), info.Name()))

	if err != nil {
		return e, err
	}

	defer file.Close()

	sha256er := sha256.New()

	if _, err = io.Copy(sha256er, file); err != nil {
		return e, err
	}

	// use RawStdEncoding to avoid padding
	// all sha256 hashes are the same length and all values would need padding anyway
	sha := sha256er.Sum(nil)
	base64 := base64.RawStdEncoding.EncodeToString(sha)

	e.path = path // note using normalized path, not info.Name()
	e.lastMod = info.ModTime()
	e.size = info.Size()
	e.hash = base64

	return e, nil
}

// Path path of the file.
func (e Entry) Path() string {
	return e.path
}

// Size gets the size fo the file, in bytes.
func (e Entry) Size() int64 {
	return e.size
}

// LastMod gets the last modification time of the file.
func (e Entry) LastMod() time.Time {
	return e.lastMod
}

// Hash gets the hash of the file as a Base64 encoded string.
func (e Entry) Hash() string {
	return e.hash
}

// IsValid returns true if all the Entry's fields are set correctly.
func (e Entry) IsValid() bool {
	// 43 == size of base64 encoded sha 256 without padding
	return (e.path != "") && !e.lastMod.IsZero() && (e.size > 0) && (e.hash != "") && (len(e.hash) == 43)
}

// AsCsv returns the entry as a comma separate string.
func (e Entry) AsCsv() string {
	return fmt.Sprintf("%s,%d,%d,%s", e.path, e.lastMod.Unix(), e.size, e.hash)
}

func (e Entry) String() string {
	return fmt.Sprintf("{path: '%s', lastMod: %s, size: %s, hash: %s}", e.path, humanize.Time(e.lastMod), humanize.Bytes(uint64(e.size)), e.hash)
}
