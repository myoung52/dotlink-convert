package main

import (
	"fmt"
	"sort"
	"strings"
)

// duplicatePaths reports an error naming every link path that appears
// more than once. A repeated path means one entry silently overwrites
// the other wherever the output actually gets used (ln just replaces
// the symlink, and a manifest map has no room for two targets), so
// it's worth catching before either format gets written out.
func duplicatePaths(links []Link) error {
	seen := make(map[string]int, len(links))
	var dupes []string
	for _, l := range links {
		seen[l.Path]++
		if seen[l.Path] == 2 {
			dupes = append(dupes, l.Path)
		}
	}
	if len(dupes) == 0 {
		return nil
	}
	sort.Strings(dupes)
	return fmt.Errorf("duplicate link path(s): %s", strings.Join(dupes, ", "))
}
