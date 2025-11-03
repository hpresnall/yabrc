package config

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"

	log "github.com/spf13/jwalterweatherman"
	"golang.org/x/text/unicode/norm"
)

// Config represents the information need to load and store an Index on the filesystem.
type Config struct {
	root        string           // base path for the files that are indexed
	savePath    string           // base path of the Index when saved to a file system
	baseName    string           // default name of Index file, without extensions
	ignoredDirs []*regexp.Regexp // list of directories to ignore when building the Index, relative to root
}

// Root returns the root directory to be used by the Index.
func (c Config) Root() string {
	return c.root
}

// SavePath returns the directory where the Index will be saved.
func (c Config) SavePath() string {
	return c.savePath
}

// BaseName returns the base name that will be used when saving the Index.
func (c Config) BaseName() string {
	return c.baseName
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

func new(root string, savePath string, baseName string, possibleRegexes []string) (Config, error) {
	var c Config

	if root == "" {
		return c, errors.New("'root' must be defined")
	}

	// change Windows \ to /
	root = strings.Replace(root, "\\", "/", -1)

	// note Clean assumes / separator
	root = norm.NFC.String(path.Clean(root))

	// add trailing / to root so Index Entries do not start with /
	// if !strings.HasSuffix(root, "/") {
	// 	root += "/"
	// }

	if baseName == "" {
		return c, errors.New("'baseName' must be defined")
	}

	// change Windows \ to /
	savePath = strings.Replace(savePath, "\\", "/", -1)

	var ignoredDirs []*regexp.Regexp

	for _, possibleRegex := range possibleRegexes {
		possibleRegex = strings.TrimSpace(possibleRegex)

		if possibleRegex == "" {
			continue
		}

		re, err := regexp.Compile(norm.NFC.String(possibleRegex))

		if err != nil {
			return c, err
		}

		ignoredDirs = append(ignoredDirs, re)
	}

	c.root = root
	c.savePath = savePath
	c.baseName = baseName
	c.ignoredDirs = ignoredDirs

	return c, nil
}
