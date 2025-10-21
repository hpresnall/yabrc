package index

import (
	"fmt"
	"path"

	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/config"
	"github.com/hpresnall/yabrc/file"
)

// Load loads the index defined by the given Config plus an (optional) identifier extension (e.g. _current, _known, etc).
func Load(config config.Config, ext string) (idx *Index, err error) {
	indexFile := GetPath(config, ext)

	idx, err = load(indexFile)

	if err != nil {
		return idx, fmt.Errorf("cannot read index from '%s': %v", indexFile, err)
	}

	log.INFO.Printf("'%s'=%v\n", indexFile, idx)

	return idx, nil
}

// Store stores the given index using Config.SavePath(), Config.BaseName()
// and an (optional) identifier extension (e.g. _current, _known, etc) as the file name.
func Store(idx *Index, config config.Config, ext string) error {
	err := file.GetFs().MkdirAll(config.SavePath(), 0755)

	if err != nil {
		return fmt.Errorf("cannot create directory '%s': %v", config.SavePath(), err)
	}

	indexFile := GetPath(config, ext)

	err = idx.Store(indexFile)

	if err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	return nil
}

// GetPath returns the filename that will be used by StoreIndex().
// It outputs Config.SavePath() + "/" + Config.BaseName() + ext.
func GetPath(config config.Config, ext string) string {
	return path.Join(config.SavePath(), config.BaseName()+ext)
}
