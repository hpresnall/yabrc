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
// It links the index file system into Viper, saves 'config.yaml'
// from the given string and then loads that file into a Config.
// The function returned is for test teardown and should be called via defer.
func ForTest(t *testing.T) (Config, func()) {
	teardown := setupViperForTest()

	configString := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: .*ignored.*`
	test.MakeFile(t, TestFile, configString, 0644)
	// if MakeFile fails, index fs probably will not be cleaned up

	c, err := Load(TestFile)

	if err != nil {
		teardown()
		t.Fatal("should be able to load config", err)
	}

	return c, teardown
}

func setupViperForTest() func() {
	fsTeardown := test.SetupTestFs()

	viperHook = func(v *viper.Viper) {
		v.SetFs(file.GetFs())
	}

	return func() {
		fsTeardown()
		viperHook = defaultViperHook
	}
}
