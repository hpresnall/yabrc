package util

import (
	"sort"
	"time"

	humanize "github.com/dustin/go-humanize"
	log "github.com/spf13/jwalterweatherman"

	"github.com/hpresnall/yabrc/index"
)

// Compare examines the Entries in the given Indexes and returns true if they are all the same.
// Supports custom output on missing entries or entries that have different hashes.
func Compare(one *index.Index, two *index.Index, ignoreMissing bool) bool {
	if one == two {
		return true
	}

	if (one == nil) || (two == nil) {
		return false
	}

	// allow comparison of files in different files systems; do not check for different roots

	sortedPaths := sortPaths(one, two)
	same := true

	// no short circuit returns in this loop to ensure that callers can track all Entries via OnMissing and OnHashChange
	for _, path := range sortedPaths {
		e1, exists1 := one.Get(path)
		e2, exists2 := two.Get(path)

		if !exists1 {
			// missing from the 1st index implies a deletion; conditionally report
			if !ignoreMissing {
				OnMissing(e2, one)
				same = false
			}
			continue
		}
		if !exists2 {
			// missing from the 2nd index implies an addition; always report
			OnMissing(e1, two)
			same = false
			continue
		}
		// assume both cannot be missing

		hash1 := e1.Hash()
		hash2 := e2.Hash()

		if hash1 != hash2 {
			OnHashChange(e1, e2)
			same = false
		}
	}

	return same
}

// MissingFn is the called when an Entry is missing from the index.
type MissingFn func(index.Entry, *index.Index)

// HashFn is called when Entries do not have the same hash (i.e. they have changed).
type HashFn func(index.Entry, index.Entry)

// OnMissing is the MissingFn that will be called.
var OnMissing MissingFn

// OnHashChange is the HashFn that will be called.
var OnHashChange HashFn

func init() {
	now := time.Now() // use fixed now to prevent time updating when there is a lot of output

	// default functions for Compare
	OnMissing = func(missing index.Entry, other *index.Index) {
		log.INFO.Printf("! '%s': '%s' %s\n", missing.Path(), other.Config().Root(), outputTime(now, other.Timestamp()))
	}

	OnHashChange = func(e1 index.Entry, e2 index.Entry) {
		diff := e1.Size() - e2.Size()

		if diff != 0 {
			comparison := ">"

			if diff < 0 {
				diff = -diff
				comparison = "<"
			}

			log.INFO.Printf("%s '%s': %s %s vs %s\n", comparison, e1.Path(), humanize.Bytes(uint64(diff)), outputTime(now, e1.LastMod()), outputTime(now, e2.LastMod()))
		} else {
			log.INFO.Printf("# '%s': %s vs %s\n", e1.Path(), outputTime(now, e1.LastMod()), outputTime(now, e2.LastMod()))
		}
	}
}

func outputTime(now time.Time, t time.Time) string {
	return humanize.RelTime(t, now, "ago", "from now")
}

// sort all the Entry paths from the two indexes
// needed for better output since Go maps are not sorted
func sortPaths(one *index.Index, two *index.Index) []string {
	// unique paths via temp map
	uniquePaths := make(map[string]struct{}, one.Size()+two.Size())

	agg := func(e index.Entry) {
		uniquePaths[e.Path()] = struct{}{}
	}

	one.ForEach(agg)
	two.ForEach(agg)

	// put paths back into a new slice
	sortedPaths := make([]string, len(uniquePaths))
	n := 0

	for path := range uniquePaths {
		sortedPaths[n] = path
		n++
	}

	sort.Strings(sortedPaths)

	return sortedPaths
}
