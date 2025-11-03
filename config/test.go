//go:build !release
// +build !release

package config

import (
	"testing"

	"github.com/spf13/viper"

	"github.com/hpresnall/yabrc/file"
	"github.com/hpresnall/yabrc/test"
)

// TestFile is the name of the config file used for testing.
var TestFile = "config.yaml"

// ForTest sets up a fake config file for use in tests.
// It links the test file system into Viper, saves 'config.yaml'
// from the given string and then loads that file into a Config.
func ForTest(t *testing.T) Config {
	configString := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: .*ignored.*`

	c, _ := FromString(t, configString)

	return c
}

// calls setupTestFs()
// links the index file system into Viper
// creates 'config.yaml' from the given string and loads it
func FromString(t *testing.T, configString string) (Config, error) {
	setupViperForTest(t)

	test.MakeFile(t, TestFile, configString, 0644)

	c, err := Load(TestFile)

	if err != nil {
		return Config{}, err
	}

	return c, err
}

func setupViperForTest(t *testing.T) {
	test.SetupTestFs(t)

	viperHook = func(v *viper.Viper) {
		v.SetFs(file.GetFs())
	}

	t.Cleanup(func() {
		viperHook = defaultViperHook
	})
}
