package util

import (
	"fmt"
	"path"

	"github.com/hpresnall/yabrc/index"
	log "github.com/spf13/jwalterweatherman"
)

// LoadConfig loads the Config defined by the given file.
func LoadConfig(configFile string) (index.Config, error) {
	log.INFO.Printf("loading Config from '%s'\n", configFile)

	config, err := index.NewConfig(configFile)

	if err != nil {
		return config, fmt.Errorf("cannot read config file '%s': %v", configFile, err)
	}

	log.INFO.Printf("'%s'=%s\n", configFile, config)

	return config, nil
}

// LoadIndex loads the index defined by the given Config plus an (optional) identifier extension (e.g. _current, _known, etc).
func LoadIndex(config index.Config, ext string) (idx *index.Index, err error) { // named returns here b/c gometalinter was erroring
	indexFile := GetIndexFile(config, ext)

	idx, err = index.Load(indexFile)

	if err != nil {
		return idx, fmt.Errorf("cannot read index from '%s': %v", indexFile, err)
	}

	log.INFO.Printf("'%s'=%v\n", indexFile, idx)

	return idx, nil
}

// StoreIndex stores the given index using Config.SavePath(), Config.BaseName()
// and an (optional) identifier extension (e.g. _current, _known, etc) as the file name.
func StoreIndex(idx *index.Index, config index.Config, ext string) error {
	err := index.GetIndexFs().MkdirAll(config.SavePath(), 0755)

	if err != nil {
		return fmt.Errorf("cannot create directory '%s': %v", config.SavePath(), err)
	}

	indexFile := GetIndexFile(config, ext)

	err = idx.Store(indexFile)

	if err != nil {
		return fmt.Errorf("cannot save index to '%s': %v", indexFile, err)
	}

	return nil
}

// GetIndexFile returns the filename that will be used by StoreIndex().
// It outputs Config.SavePath() + "/" + Config.BaseName() + ext.
func GetIndexFile(config index.Config, ext string) string {
	return path.Join(config.SavePath(), config.BaseName()+ext)
}
