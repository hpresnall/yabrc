package index

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"errors"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	humanize "github.com/dustin/go-humanize"
	log "github.com/spf13/jwalterweatherman"
	"golang.org/x/text/unicode/norm"

	"github.com/hpresnall/yabrc/file"
)

// Index stores data for all files under a particular root directory.
type Index struct {
	root    string // root path for all data
	rootLen int

	timestamp time.Time

	// index is a map of the file path to a row of data
	data map[string]Entry
}

// New creates a new, empty index rooted at the given path. This path _must_ use / as the separator.
func New(root string) (*Index, error) {
	if root == "" {
		return &Index{}, errors.New("root cannot be empty")
	}
	// note Clean assumes / separator; assume the same here
	root = norm.NFC.String(path.Clean(root))

	// add trailing / to root so Index Entries do not start with /; see Add()
	if !strings.HasSuffix(root, "/") {
		root += "/"
	}

	// truncate time for Load & Store comparisions since it will only be stored as Unix time
	return &Index{root: root, rootLen: utf8.RuneCountInString(root), timestamp: time.Now().Truncate(time.Second), data: make(map[string]Entry)}, nil
}

// load loads an existing index from the given path.
// Bad data in the Index will be logged but processing will continue to the end of the file.
func load(path string) (*Index, error) {
	// fix \ to /
	path = strings.Replace(path, "\\", "/", -1)

	log.DEBUG.Printf("loading Index from '%s'", path)

	var idx *Index

	in, err := file.GetFs().Open(path)

	if err != nil {
		return idx, err
	}

	defer in.Close()

	gz, err := gzip.NewReader(in)

	if err != nil {
		return idx, err
	}

	defer gz.Close()

	// use Scanner rather than csv.Reader
	// the latter does not skip blank lines or have any mechanism to tell you that fields are missing other than errors
	r := bufio.NewScanner(gz)
	r.Split(bufio.ScanLines)

	readHeader := false
	n := 0

	for r.Scan() {
		n++
		fields := strings.Split(r.Text(), ",")
		originalFields := make([]string, len(fields))

		// skip blank lines
		// note lines with all commas get parsed to single empty string
		blanks := 0

		for n := range fields {
			originalFields[n] = fields[n]
			fields[n] = strings.TrimSpace(fields[n])

			if fields[n] == "" {
				blanks++
			}
		}

		log.TRACE.Printf("%d: '%s' => %q, %d blanks\n", n, r.Text(), fields, blanks)

		if blanks == len(fields) {
			log.TRACE.Printf("%d: skipping blank line\n", n)
			continue
		}

		if !readHeader {
			idx, err = New(fields[0])

			if err != nil {
				return idx, fmt.Errorf("%d: header '%s' must define a root path", n, r.Text())
			}

			rawTime, err := strconv.ParseInt(fields[1], 10, 64)

			if err != nil {
				return idx, fmt.Errorf("%d: header '%s' must include integer timestamp", n, r.Text())
			}

			idx.timestamp = time.Unix(rawTime, 0)

			readHeader = true
			continue
		}

		entryPath := fields[0]
		i := 1

		// rather than add quotes for paths with , in Store(), just concat here and adjust other indexes
		// use the original field since there may be spaces around the comma
		for ; i < (len(fields) - 3); i++ {
			entryPath = entryPath + "," + originalFields[i]
		}
		entryPath = norm.NFC.String(entryPath) // normalize paths that contain Unicode combining characters to a single character

		rawTime, err := strconv.ParseInt(fields[i], 10, 64)

		if err != nil {
			log.WARN.Printf("%d: skipping line '%s'; '%s' must be a Unix time value", n, r.Text(), fields[i])
			continue
		}

		lastMod := time.Unix(rawTime, 0)

		size, err := strconv.ParseInt(fields[i+1], 10, 64)

		if err != nil {
			log.WARN.Printf("%d: skipping line '%s'; %s must be an integer", n, r.Text(), fields[i+1])
			continue
		}

		entry := Entry{path: entryPath, lastMod: lastMod, size: size, hash: fields[i+2]}
		idx.data[entryPath] = entry

		// all fields parsed ok; log extra commas, but do not mark as a error
		if i > 1 {
			log.TRACE.Printf("%d: line '%s' had too many commas", n, r.Text())
		}

		log.TRACE.Printf("%v: added %v\n", idx, entry)
	}

	if n == 0 {
		return idx, errors.New("no data loaded from file")
	}

	return idx, nil
}

// Store writes the index to the file system at the given path.
// This path should be a file. Callers are responsible for naming the file.
func (idx *Index) Store(path string) error {
	if idx.root == "" {
		return errors.New("cannot store Index with empty root")
	}

	log.DEBUG.Printf("storing Index to '%s'", path)

	out, err := file.GetFs().Create(path)

	if err != nil {
		return err
	}

	defer out.Close()

	// gzip the file to save space and for minor obfuscation / edit protection
	gz := gzip.NewWriter(out)

	// not using csv.Writer since data needs to be converted to strings anyway, Sprintf is easier

	// format is root,time on its own line, followed by CSV output for each Entry
	_, err = gz.Write([]byte(fmt.Sprintf("%s,%d\n", idx.Root(), idx.timestamp.Unix())))

	if err != nil {
		return err
	}

	for _, entry := range idx.data {
		csv := entry.AsCsv()

		log.TRACE.Println("writing", csv)

		_, err = gz.Write([]byte(csv))

		if err != nil {
			return err
		}

		_, err = gz.Write([]byte("\n"))

		if err != nil {
			return err
		}
	}

	if err = gz.Flush(); err != nil {
		return err
	}

	if err = gz.Close(); err != nil {
		return err
	}

	// ensure file is written
	return out.Sync()
}

// Root returns the base path for the Entries in the index.
// Entry.Path should be combined with this path to get the absolute path of a file.
func (idx *Index) Root() string {
	return idx.root
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

	if !strings.HasPrefix(path, idx.Root()) {
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
		if !strings.HasPrefix(entry.path, idx.Root()) {
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

// GetRelativePath returns the given path without the Index root.
func (idx *Index) GetRelativePath(path string) string {
	path = norm.NFC.String(path)
	if strings.HasPrefix(path, idx.Root()) {
		pathFromRoot := string([]rune(path)[idx.rootLen:])

		return pathFromRoot
	}

	return path
}

// Get the entry for the given path.
// This path _must be_ relative to the index root; see GetRelativePath()
func (idx *Index) Get(path string) (Entry, bool) {
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
	return fmt.Sprintf("{root: '%s', timestamp: %s, size: %d}", idx.Root(), humanize.Time(idx.Timestamp()), idx.Size())
}

// StringWithEntries returns the Index as JSON string that contains all of the Entries.
func (idx *Index) StringWithEntries() string {
	var buffer bytes.Buffer

	size := idx.Size()

	// very roughly index output + size * entry output
	buffer.Grow(75 + size*150)

	buffer.WriteString("{\"root\": \"")
	buffer.WriteString(idx.Root())
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
