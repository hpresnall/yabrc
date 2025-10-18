package index

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	log "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
	"golang.org/x/text/unicode/norm"
)

// Config represents the information need to load and store an Index on the filesystem.
type Config struct {
	root        string           // Index root
	baseName    string           // default name of Index file, without extensions
	savePath    string           // path of Index file
	ignoredDirs []*regexp.Regexp // list of directories to ignore when building the Index
}

// NewConfig creates a Config from the given file.
func NewConfig(configFile string) (Config, error) {
	var config Config

	v := viper.New()
	v.SetConfigFile(configFile)
	ConfigViperHook(v)

	if err := v.ReadInConfig(); err != nil {
		return config, err
	}

	root := v.GetString("root")

	if root == "" {
		return config, errors.New("'root' must be defined")
	}

	// change Windows \ to /
	config.root = strings.Replace(root, "\\", "/", -1)

	baseName := v.GetString("baseName")

	if baseName == "" {
		return config, errors.New("'baseName' must be defined")
	}

	config.baseName = baseName

	savePath := v.GetString("savePath")

	// default to the same directory as the config file
	if savePath == "" {
		// path.Dir assumes / for separator; Replace() first to ensure it works
		savePath = path.Dir(strings.Replace(configFile, "\\", "/", -1))
		log.DEBUG.Printf("set empty 'savePath' to '%s'", savePath)
	} else {
		savePath = strings.Replace(savePath, "\\", "/", -1)
	}

	config.savePath = savePath

	possibleRegexs := v.GetStringSlice("ignoredDirs")
	var ignoredDirs []*regexp.Regexp

	for _, possibleRegex := range possibleRegexs {
		possibleRegex = strings.TrimSpace(possibleRegex)

		if possibleRegex == "" {
			continue
		}

		re, err := regexp.Compile(norm.NFC.String(possibleRegex))

		if err != nil {
			return config, err
		}

		ignoredDirs = append(ignoredDirs, re)
	}

	config.ignoredDirs = ignoredDirs

	return config, nil
}

// Root returns the root directory to be used by the Index.
func (c Config) Root() string {
	return c.root
}

// BaseName returns the base name that will be used when saving the Index.
func (c Config) BaseName() string {
	return c.baseName
}

// SavePath returns the directory where the Index will be saved.
func (c Config) SavePath() string {
	return c.savePath
}

// IgnoreDir returns true if the given directory matches any of the ignored directory regular expressions.
func (c Config) IgnoreDir(dir string) bool {
	dir = norm.NFC.String(dir) // normalize to match compiled regexes

	for _, re := range c.ignoredDirs {
		if re.MatchString(dir) {
			log.TRACE.Println(dir, "matches", re)
			return true
		}
	}

	return false
}

// Formats the config as a String.
func (c Config) String() string {
	ignoredStrings := make([]string, len(c.ignoredDirs))

	for i, re := range c.ignoredDirs {
		ignoredStrings[i] = re.String()
	}

	return fmt.Sprintf("{root: '%s', baseName: '%s', savePath: '%s', ignoredDirs: [ %s ]}", c.root, c.baseName, c.savePath, strings.Join(ignoredStrings, ", "))
}

// ConfigViperHook is a hook function meant for testing.
// This is called after the Viper instance is created but before the Config is loaded from the file system.
var ConfigViperHook func(v *viper.Viper)

func init() {
	ConfigViperHook = func(v *viper.Viper) {}
}
