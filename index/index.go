package index

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	humanize "github.com/dustin/go-humanize"
	log "github.com/spf13/jwalterweatherman"
	"golang.org/x/text/unicode/norm"

	"github.com/hpresnall/yabrc/config"
)

// Index stores data for all files under a Config's root directory.
type Index struct {
	config    *config.Config
	rootLen   int              // length of Config.root for comparisons
	timestamp time.Time        // time, in epoch seconds, when the index was buil
	data      map[string]Entry // the index is a map of the file path to a row of data

	rootWithSlash string
}

// New creates a new, empty index using the given Config.
func New(config *config.Config) (*Index, error) {
	if config == nil {
		return nil, errors.New("cannot create an index with a nil Config")
	}

	return &Index{
		config:  config,
		rootLen: utf8.RuneCountInString(config.Root()) + 1, // add one to rootLen to avoid Entries starting with /
		// truncate time for Load & Store comparisions since it will only be stored as Unix time
		timestamp:     time.Now().Truncate(time.Second),
		data:          make(map[string]Entry),
		rootWithSlash: config.Root() + "/",
	}, nil
}

func (idx *Index) Config() *config.Config {
	return idx.config
}

// Timestamp returns the datetime when this index was created.
func (idx *Index) Timestamp() time.Time {
	return idx.timestamp
}

// Size returns the number of Entries in the index.
func (idx *Index) Size() int {
	return len(idx.data)
}

// Add the given file to the index after parsing it to a new Entry.
// The given path must include the index's root.
func (idx *Index) Add(path string, info os.FileInfo) error {
	// skip 0 byte files
	if (info != nil) && (info.Size() <= 0) {
		log.DEBUG.Printf("%v: skipped zero byte file '%s'\n", idx, path)
		return nil
	}

	// ensure Windows \ are changed to /
	path = norm.NFC.String(strings.Replace(path, "\\", "/", -1))

	if !strings.HasPrefix(path, idx.config.Root()) {
		return fmt.Errorf("%v: path '%s' does not start with root", idx, path)
	}

	entry, err := buildEntry(path, info)

	if err != nil {
		return err
	}

	idx.addEntryToMap(entry)

	return nil
}

// AddEntry adds the given Entry to the index.
func (idx *Index) AddEntry(entry Entry) error {
	if entry.IsValid() {
		if !strings.HasPrefix(entry.path, idx.rootWithSlash) {
			// add the entry without changing its path
			idx.data[entry.path] = entry

			log.TRACE.Printf("%v: added %v\n", idx, entry)
		} else {
			idx.addEntryToMap(entry)
		}

		return nil
	}

	return fmt.Errorf("%v: cannot add invalid entry: '%v'", idx, entry)
}

func (idx *Index) addEntryToMap(entry Entry) {
	// store with relative path to save space / memory
	pathFromRoot := string([]rune(entry.path)[idx.rootLen:])
	entry.path = pathFromRoot // note this _does not_ update the original Entry since it is not a pointer

	idx.data[pathFromRoot] = entry

	log.TRACE.Printf("%v: added %v\n", idx, entry)
}

// Get the entry for the given path.
// This path _must be_ relative to the index root; see GetRelativePath()
func (idx *Index) Get(path string) (Entry, bool) {
	path = norm.NFC.String(path)

	if strings.HasPrefix(path, idx.rootWithSlash) {
		path = string([]rune(path)[idx.rootLen:])
	}

	entry, exists := idx.data[path]
	return entry, exists
}

// ForEach Entry in the index, execute the given function.
func (idx *Index) ForEach(f func(Entry)) {
	for _, entry := range idx.data {
		f(entry)
	}
}

func (idx *Index) String() string {
	return fmt.Sprintf("{root: '%s', timestamp: %s, size: %d}", idx.config.Root(), humanize.Time(idx.Timestamp()), idx.Size())
}

// StringWithEntries returns the Index as JSON string that contains all of the Entries.
func (idx *Index) StringWithEntries() string {
	var buffer bytes.Buffer

	size := idx.Size()

	// very roughly index output + size * entry output
	buffer.Grow(75 + size*150)

	buffer.WriteString("{\"root\": \"")
	buffer.WriteString(idx.config.Root())
	buffer.WriteString("\", \"timestamp\": ")
	buffer.WriteString(strconv.FormatInt(idx.Timestamp().Unix(), 10))
	buffer.WriteString(", \"size\": ")
	buffer.WriteString(strconv.Itoa(size))
	buffer.WriteString(", \"entries\": [")

	n := 1

	idx.ForEach(func(e Entry) {
		buffer.WriteString("{\"path\": \"")
		buffer.WriteString(e.Path())
		buffer.WriteString("\", \"lastMod\": ")
		buffer.WriteString(strconv.FormatInt(e.LastMod().Unix(), 10))
		buffer.WriteString(", \"size\": ")
		buffer.WriteString(strconv.FormatInt(e.Size(), 10))
		buffer.WriteString(", \"hash\": \"")
		buffer.WriteString(e.Hash())
		buffer.WriteString("\"}")

		if n < size {
			buffer.WriteString(", ")
		}

		n++
	})

	buffer.WriteString("]}")

	return buffer.String()
}
