package config

import (
	"testing"
)

func TestConfig(t *testing.T) {
	// spaces in ignoredDirs to ensure they are trimmed; also ensure empty strings are not stored
	config := `root: testRoot
baseName: testBaseName
savePath: testSavePath
ignoredDirs: [' ignored.* ', '', ' test.*']
`
	c, err := FromString(t, config)

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
}

func TestEmptyConfig(t *testing.T) {
	_, err := FromString(t, "")

	if err == nil {
		t.Error("should not be able to load empty config", err)
	}
}

func TestInvalidConfig(t *testing.T) {
	_, err := FromString(t, "missing")

	if err == nil {
		t.Error("should not create a config from missing file")
	}
}

func TestConfigWithNoRoot(t *testing.T) {
	_, err := FromString(t, "baseName: testBaseName")

	if err == nil {
		t.Error("should not be able to load config with no root", err)
	}
}

func TestConfigWithNoBaseName(t *testing.T) {
	_, err := FromString(t, "root: testRoot")

	if err == nil {
		t.Error("should not be able to load config with no baseName", err)
	}
}

func TestConfigWithNoSavePath(t *testing.T) {
	config := `root: testRoot
baseName: testBaseName
`
	c, err := FromString(t, config)

	if err != nil {
		t.Fatal("cannot load config without savePath", err)
	}

	if c.SavePath() != "." {
		t.Error("savePath should default to the current working dir not", c.SavePath())
	}
}

func TestConfigWithNoIgnoredDirs(t *testing.T) {
	// config should load with an empty array
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
	c, err := FromString(t, config)

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
	c, err := FromString(t, config)

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
	_, err := FromString(t, config)

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

// for coverage
func TestForTest(t *testing.T) {
	config := ForTest(t)

	if config.root != "testRoot" {
		t.Fatal("should have created valid config")
	}
}
