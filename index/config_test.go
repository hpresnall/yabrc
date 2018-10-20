package index

import (
	"testing"

	"github.com/spf13/afero"
	log "github.com/spf13/jwalterweatherman"
	"github.com/spf13/viper"
)

func TestConfig(t *testing.T) {
	// spaces in ignoredDirs to ensure they are trimmed; extra commas to ensure empty strings are not stored
	config := `root=testRoot
baseName=testBaseName
savePath=testSavePath
ignoredDirs=,,ignored.*, ,  test.* ,,
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
		t.Error("should have 2 regexps in ignoredDirs")
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
	_, err := NewConfig("")

	if err == nil {
		t.Error("should not be able to load empty config", err)
	}
}

func TestMissingConfig(t *testing.T) {
	_, err := NewConfig("missing")

	if err == nil {
		t.Error("should not create a config from missing file")
	}
}

func TestConfigWithNoRoot(t *testing.T) {
	_, teardown, err := newConfigFromString(t, "baseName=testBaseName")
	defer teardown()

	if err == nil {
		t.Error("should not be able to load config with no root", err)
	}
}

func TestConfigWithNoBaseName(t *testing.T) {
	_, teardown, err := newConfigFromString(t, "root=testRoot")
	defer teardown()

	if err == nil {
		t.Error("should not be able to load config with no baseName", err)
	}
}

func TestConfigWithNoSavePath(t *testing.T) {
	config := `root=testRoot
baseName=testBaseName
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
	config := `root=testRoot
baseName=testBaseName
`
	c, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err != nil {
		t.Fatal("cannot load config", err)
	}

	if c.ignoredDirs != nil {
		t.Fatal("ignoredDirs should be nil")
	}

	if len(c.ignoredDirs) != 0 {
		t.Fatal("ignoredDirs shoiuld be empty")
	}
}

func TestConfigWithBadIgnoredDirs(t *testing.T) {
	config := `root=testRoot
baseName=testBaseName
ignoredDirs=[
`
	_, teardown, err := newConfigFromString(t, config)
	defer teardown()

	if err == nil {
		t.Fatal("should not load config with bad ignoredDirs", err)
	}
}

// calls setupFs()
// links the index file system into Viper
// creates 'config.properties' from the given string and loads it
func newConfigFromString(t *testing.T, configString string) (Config, func(), error) {
	testFs, testFsTeardown := setupTestFs()

	err := afero.WriteFile(testFs, "config.properties", []byte(configString), 0644)

	if err != nil {
		// Fatal stops the goroutine before the caller can defer the teardown function
		// run it manually now
		testFsTeardown()
		ConfigViperHook = func(v *viper.Viper) {}

		t.Fatal("cannot make file", "config.properties", err)
	}

	ConfigViperHook = func(v *viper.Viper) {
		v.SetFs(GetIndexFs())
	}

	log.DEBUG.Printf("loading config from '%s'\n", configString)

	c, err := NewConfig("config.properties")

	return c, func() {
		testFsTeardown()
		ConfigViperHook = func(v *viper.Viper) {}
	}, err
}
