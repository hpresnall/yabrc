//go:build !release
// +build !release

package index

import (
	"testing"

	"github.com/hpresnall/yabrc/config"
)

func ForTest(t *testing.T) *Index {
	config := config.ForTest(t)
	idx, _ := New(&config)

	return idx
}
