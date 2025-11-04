package config

import (
	"fmt"
	"path"
	"strings"

	log "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

// Load loads the Config defined by the given file.
func Load(configFile string) (Config, error) {
	log.INFO.Printf("loading Config from '%s'\n", configFile)

	v := viper.New()
	v.SetConfigFile(configFile)
	viperHook(v)

	if err := v.ReadInConfig(); err != nil {
		return Config{}, err
	}

	root := v.GetString("root")
	savePath := v.GetString("savePath")
	baseName := v.GetString("baseName")
	regexes := v.GetStringSlice("ignoredDirs")

	// default to the same directory as the config file
	if savePath == "" {
		// path.Dir assumes / for separator; Replace() first to ensure it works
		savePath = path.Dir(strings.Replace(configFile, "\\", "/", -1))
		log.DEBUG.Printf("set empty 'savePath' to '%s'", savePath)
	}

	config, err := new(root, savePath, baseName, regexes)

	if err != nil {
		return config, fmt.Errorf("cannot read config file '%s': %v", configFile, err)
	}

	log.INFO.Printf("'%s'=%s\n", configFile, config)

	return config, nil
}

// ViperHook is a hook function meant for testing.
// This is called after the Viper instance is created but before the Config is loaded from the file system.
var viperHook func(v *viper.Viper)
var defaultViperHook = func(v *viper.Viper) {}

func init() {
	viperHook = defaultViperHook
}
