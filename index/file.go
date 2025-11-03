package index

import (
	"bufio"
	"compress/gzip"
	"errors"
	"fmt"
	"path"
	"strconv"
	"strings"
	"time"

	log "github.com/spf13/jwalterweatherman"
	"golang.org/x/text/unicode/norm"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
)

// Load loads the index defined by the given Config plus an (optional) identifier extension (e.g. _current, _known, etc).
func Load(config *config.Config, ext string) (*Index, error) {
	idx, err := New(config)

	if err != nil {
		return idx, nil
	}

	file := idx.GetFile(ext)

	log.DEBUG.Printf("loading Index from '%s'", file)

	err = load(idx, file)

	if err != nil {
		return idx, fmt.Errorf("cannot read index from '%s': %v", file, err)
	}

	log.INFO.Printf("'%s'=%v\n", file, idx)

	return idx, nil
}

// load loads an existing index from the given path.
// Bad data in the Index will be logged but processing will continue to the end of the file.
func load(idx *Index, path string) error {
	in, err := file.GetFs().Open(path)

	if err != nil {
		return err
	}

	defer in.Close()

	gz, err := gzip.NewReader(in)

	if err != nil {
		return err
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
			// support old indexes that wrote rootWithSlash to header
			if (fields[0] != idx.Config().Root()) && (fields[0] != idx.rootWithSlash) {
				return fmt.Errorf("%d: header '%s' must define a root path that matches Config.Root '%s'", n, r.Text(), idx.Config().Root())
			}

			rawTime, err := strconv.ParseInt(fields[1], 10, 64)

			if err != nil {
				return fmt.Errorf("%d: header '%s' must include integer timestamp", n, r.Text())
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
		return errors.New("no data loaded from file")
	}

	return nil
}

// Store writes the index to the file system with the given extension.
func (idx *Index) Store(ext string) error {
	if len(idx.data) == 0 {
		return fmt.Errorf("cannnot store an empty index")
	}

	indexFile := idx.GetFile(ext)

	log.DEBUG.Printf("storing Index to '%s'", indexFile)

	err := file.GetFs().MkdirAll(idx.Config().SavePath(), 0755)

	if err != nil {
		return fmt.Errorf("cannot create directory '%s': %v", idx.Config().SavePath(), err)
	}

	out, err := file.GetFs().Create(indexFile)

	if err != nil {
		return err
	}

	defer out.Close()

	// gzip the file to save space and for minor obfuscation / edit protection
	gz := gzip.NewWriter(out)

	// not using csv.Writer since data needs to be converted to strings anyway, Sprintf is easier

	// format is root,time on its own line, followed by CSV output for each Entry
	_, err = gz.Write([]byte(fmt.Sprintf("%s,%d\n", idx.Config().Root(), idx.timestamp.Unix())))

	if err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	for _, entry := range idx.data {
		csv := entry.AsCsv()

		log.TRACE.Println("writing", csv)

		_, err = gz.Write([]byte(csv))

		if err != nil {
			return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
		}

		_, err = gz.Write([]byte("\n"))

		if err != nil {
			return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
		}
	}

	if err = gz.Flush(); err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	if err = gz.Close(); err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	// ensure file is written
	if err = out.Sync(); err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	return nil
}

// GetFile returns the filename that will be used by Store().
func (idx *Index) GetFile(ext string) string {
	return GetIndexFile(idx.Config(), ext)
}

// GetFile returns the filename that will be used by Store().
// It outputs Config.SavePath() + "/" + Config.BaseName() + ext.
func GetIndexFile(config *config.Config, ext string) string {
	return path.Join(config.SavePath(), config.BaseName()+ext)
}
