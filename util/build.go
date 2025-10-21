package util

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"time"

	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/afero"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/index"
)

// BuildIndex creates an Index by walking the file system from Config.Root().
// If an existing Index is passed in, only new & updated files will be scanned. Other files will use
// the existing Index's Entries.
func BuildIndex(config config.Config, existingIdx *index.Index) (*index.Index, error) {
	idx, err := index.New(config.Root())

	if err != nil {
		return idx, err
	}

	log.INFO.Printf("building index for '%s'\n", idx.Root())

	start := time.Now()

	dirCount := 0
	zeroCount := 0
	nonCount := 0
	metadataCount := 0
	hashedCount := 0
	hashedBytes := int64(0)
	existingCount := 0
	errCount := 0
	skippedBytes := int64(0)

	err = afero.Walk(file.GetFs(), idx.Root(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			errCount++
			log.WARN.Println("error reading file:", err.Error())
			return nil
		}

		if info.IsDir() {
			if config.IgnoreDir(path) {
				log.DEBUG.Printf("skipping dir '%s'", path)
				return filepath.SkipDir
			}
			log.DEBUG.Printf("indexing dir '%s'", path)

			dirCount++
			return nil
		} else if strings.HasSuffix(path, "desktop.ini") || strings.HasSuffix(path, ".DS_Store") {
			// do not index file manager metadata since these can change frequently
			metadataCount++
			return nil
		}

		if (info.Mode() & os.ModeType) != 0 {
			log.WARN.Printf("skipping non-file '%s' (%s)\n", path, info.Mode())
			nonCount++
			return nil
		}

		if info.Size() <= 0 {
			zeroCount++
			return nil
		}

		if existingIdx != nil {
			relativePath := idx.GetRelativePath(strings.Replace(path, "\\", "/", -1))
			entry, exists := existingIdx.Get(relativePath)

			// Entry.LastMod() stored as Unix time
			infoTime := info.ModTime().Truncate(time.Second)

			// only add existing entry if the file was created afterwards or the sizes has changed
			if exists &&
				(entry.Size() == info.Size()) &&
				(infoTime.Before(entry.LastMod()) || infoTime.Equal(entry.LastMod())) {
				if log.GetLogThreshold() == log.LevelTrace {
					log.TRACE.Printf("skipping '%s': '%v' vs '%v' & '%d' vs '%d'", relativePath, info.ModTime(), entry.LastMod(), info.Size(), entry.Size())
				}
				existingCount++
				skippedBytes += info.Size()
				err = idx.AddEntry(entry)
			} else {
				if log.GetLogThreshold() <= log.LevelDebug {
					log.DEBUG.Printf("rescanning '%s': '%v' vs '%v' & '%d' vs '%d'", relativePath, info.ModTime(), entry.LastMod(), info.Size(), entry.Size())
				}
				hashedCount++
				hashedBytes += info.Size()
				err = idx.Add(path, info)
			}
		} else {
			if log.GetLogThreshold() <= log.LevelDebug {
				log.DEBUG.Printf("adding '%s': '%v' & '%d'", path, info.ModTime(), info.Size())
			}
			hashedCount++
			hashedBytes += info.Size()
			err = idx.Add(path, info)
		}

		if err != nil {
			errCount++
			log.ERROR.Printf("cannot add '%s' to database: %v\n", path, err)
		}

		return nil
	})

	// index is truly empty, not just empty because all the files could not be read
	if (idx.Size() == 0) && (errCount > 0) {
		err = errors.New("no files successfully read from '" + idx.Root() + "'")
	}

	d := time.Since(start)
	skippedCount := zeroCount + nonCount + existingCount + metadataCount

	dRounded := d.Round(time.Second)

	// display ms as ms, otherwise show rounded
	if dRounded == 0 {
		dRounded = d
	}

	log.INFO.Printf("%v indexed %s in %v\n", idx, humanize.Bytes(uint64(hashedBytes)), dRounded)
	log.INFO.Printf("%d directories, %d files hashed, %d errors; %.f files/sec; %s/sec\n", dirCount, hashedCount, errCount, float64(hashedCount)/d.Seconds(), humanize.Bytes(uint64(float64(hashedBytes)/d.Seconds())))

	if skippedCount > 0 {
		log.INFO.Printf("%d skipped (%s); %d not changed, %d zero byte, %d dir metadata, %d non-file", skippedCount, humanize.Bytes(uint64(skippedBytes)), existingCount, zeroCount, metadataCount, nonCount)
	}

	// return err from filepath.Walk(), if any
	return idx, err
}
