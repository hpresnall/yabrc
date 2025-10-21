package config

import (
	"testing"

	"github.com/spf13/afero"

	"github.com/hpresnall/yabrc/file"
)

func TestConfig(t *testing.T) {
	// spaces in ignoredDirs to ensure they are trimmed; also ensure empty strings are not stored
	config := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: [' ignored.* ', '', ' test.*']
`
	c, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err != nil {
		t.Fatal("cannot load config", err)
	}

	if c.Root() != "testRoot" {
		t.Errorf("%s should be '%s', not '%s'", "root", "testRoot", c.Root())
	}

	if c.BaseName() != "testBaseName" {
		t.Errorf("%s should be '%s', not '%s'", "baseName", "testBaseNAme", c.BaseName())
	}

	if c.SavePath() != "testSavePath" {
		t.Errorf("%s should be '%s', not '%s'", "testSavePath", "testSavePath", c.SavePath())
	}

	if len(c.ignoredDirs) != 2 {
		t.Fatal("should have 2 regexps in ignoredDirs")
	}

	if c.ignoredDirs[0].String() != "ignored.*" {
		t.Errorf("%s[%d] should be '%s' not '%s'", "ignoredDirs", 0, "ignored.*", c.ignoredDirs[0].String())
	}

	if c.ignoredDirs[1].String() != "test.*" {
		t.Errorf("%s[%d] should be '%s' not '%s'", "ignoredDirs", 1, "test.*", c.ignoredDirs[1].String())
	}

	if !c.IgnoreDir("test/bar") {
		t.Error("should have ignored directory 'foo/test/bar'")
	}

	if c.IgnoreDir("foo/ignore/bar") {
		t.Error("should not have ignored directory 'foo/ignore/bar'")
	}

	// coverage for String()
	s := c.String()

	if s == "" {
		t.Error("should return string")
	}
}

func TestEmptyConfig(t *testing.T) {
	_, err := new("")

	if err == nil {
		t.Error("should not be able to load empty config", err)
	}
}

func TestInvalidConfig(t *testing.T) {
	_, err := new("missing")

	if err == nil {
		t.Error("should not create a config from missing file")
	}
}

func TestConfigWithNoRoot(t *testing.T) {
	_, teardown, err := newConfigFromString(t, "baseName: testBaseName")
	defer teardown()

	if err == nil {
		t.Error("should not be able to load config with no root", err)
	}
}

func TestConfigWithNoBaseName(t *testing.T) {
	_, teardown, err := newConfigFromString(t, "root: testRoot")
	defer teardown()

	if err == nil {
		t.Error("should not be able to load config with no baseName", err)
	}
}

func TestConfigWithNoSavePath(t *testing.T) {
	config := `root: testRoot
baseName: testBaseName
`
	c, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err != nil {
		t.Fatal("cannot load config", err)
	}

	if c.SavePath() != "." {
		t.Error("savePath should default to the current working dir not", c.SavePath())
	}
}

func TestConfigWithNoIgnoredDirs(t *testing.T) {
	testNoIgnoredDirs(t, `root: testRoot
baseName: testBaseName
`)
}
func TestConfigWithEmptyIgnoredDirs(t *testing.T) {
	testNoIgnoredDirs(t, `root: testRoot
baseName: testBaseName
ignoredDirs: []
`)
}

func TestConfigWithNilIgnoredDirs(t *testing.T) {
	// config should load with an empty array
	testNoIgnoredDirs(t, `root: testRoot
baseName: testBaseName
ignoredDirs:
`)
}

func testNoIgnoredDirs(t *testing.T, config string) {
	c, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err != nil {
		t.Fatal("cannot load config", err)
	}

	if c.ignoredDirs != nil {
		t.Fatal("ignoredDirs should not be nil")
	}

	if len(c.ignoredDirs) != 0 {
		t.Fatal("ignoredDirs should be empty")
	}
}

func TestConfigWithStringIgnoredDirs(t *testing.T) {
	// spaces in ignoredDirs to ensure they are trimmed
	config := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: ' ignored.* '
`
	c, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err != nil {
		t.Fatal("cannot load config", err)
	}

	if len(c.ignoredDirs) != 1 {
		t.Fatal("should have 1 regexp in ignoredDirs")
	}

	if c.ignoredDirs[0].String() != "ignored.*" {
		t.Errorf("%s[%d] should be '%s' not '%s'", "ignoredDirs", 0, "ignored.*", c.ignoredDirs[0].String())
	}
}

func TestConfigWithInvalidIgnoredDirs(t *testing.T) {
	config := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: ' [] '
`
	_, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err == nil {
		t.Error("should not be able to load config with invalid ignoredDirs", err)
	}
}

func TestLoadMissingConfig(t *testing.T) {
	_, err := Load("missing")

	if err == nil {
		t.Error("should not be able to load config from a missing file")
	}
}

// calls setupTestFs()
// links the index file system into Viper
// creates 'config.yaml' from the given string and loads it
func newConfigFromString(t *testing.T, configString string) (Config, func(), error) {
	teardown := setupViperForTest()

	// not using file.MakeFile() because failures there would not call teardown()
	err := afero.WriteFile(file.GetFs(), TestFile, []byte(configString), 0644)

	if err != nil {
		// Fatal stops the goroutine before the caller can defer the teardown function
		// run it manually now
		teardown()
		t.Fatal("cannot make file", TestFile, err)
	}

	c, err := new(TestFile)

	return c, teardown, err
}

// for coverage
func TestConfigForTest(t *testing.T) {
	config, teardown := ForTest(t)
	defer teardown()

	if config.root != "testRoot" {
		t.Fatal("should have created valid config")
	}
}
