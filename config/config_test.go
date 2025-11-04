package config

import (
	"strings"
	"testing"
)

func TestString(t *testing.T) {
	c, err := new("testRoot", "testBaseName", "testSavePath", []string{"ignored"})

	if err != nil {
		t.Fatal("cannot create config", err)
	}

	s := c.String()

	if s == "" {
		t.Fatal("should return non-empty string")
	}

	if !strings.Contains(s, "ignored") {
		t.Fatal("should contain ignoredDirs")
	}
}

// tests in file_test.go cover all other Config.new() paths
