package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// checkResult is the outcome of checking a single link against disk.
type checkResult struct {
	Link   Link
	Status string // "ok", "missing", "not-symlink", "wrong-target", "broken", "error"
	Detail string
}

// checkLinks stats each link's path on disk and compares what it
// actually points to against what the link says it should.
func checkLinks(links []Link) []checkResult {
	results := make([]checkResult, len(links))
	for i, l := range links {
		results[i] = checkLink(l)
	}
	return results
}

func checkLink(l Link) checkResult {
	info, err := os.Lstat(l.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return checkResult{Link: l, Status: "missing"}
		}
		return checkResult{Link: l, Status: "error", Detail: err.Error()}
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return checkResult{Link: l, Status: "not-symlink"}
	}
	actual, err := os.Readlink(l.Path)
	if err != nil {
		return checkResult{Link: l, Status: "error", Detail: err.Error()}
	}
	if actual != l.Target {
		return checkResult{Link: l, Status: "wrong-target", Detail: actual}
	}

	// A relative target is resolved by the kernel relative to the
	// symlink's own directory, not the process's cwd.
	resolved := l.Target
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(filepath.Dir(l.Path), resolved)
	}
	if _, err := os.Stat(resolved); err != nil {
		if os.IsNotExist(err) {
			return checkResult{Link: l, Status: "broken"}
		}
		return checkResult{Link: l, Status: "error", Detail: err.Error()}
	}
	return checkResult{Link: l, Status: "ok"}
}

// writeCheckReport prints one line per link and reports whether every
// link came back ok.
func writeCheckReport(w io.Writer, results []checkResult) bool {
	allOK := true
	for _, r := range results {
		switch r.Status {
		case "ok":
			fmt.Fprintf(w, "ok            %s\n", r.Link.Path)
		case "missing":
			allOK = false
			fmt.Fprintf(w, "missing       %s (expected -> %s)\n", r.Link.Path, r.Link.Target)
		case "not-symlink":
			allOK = false
			fmt.Fprintf(w, "not a symlink %s (expected -> %s)\n", r.Link.Path, r.Link.Target)
		case "wrong-target":
			allOK = false
			fmt.Fprintf(w, "wrong target  %s -> %s (expected -> %s)\n", r.Link.Path, r.Detail, r.Link.Target)
		case "broken":
			allOK = false
			fmt.Fprintf(w, "broken        %s -> %s (target does not exist)\n", r.Link.Path, r.Link.Target)
		case "error":
			allOK = false
			fmt.Fprintf(w, "error         %s: %s\n", r.Link.Path, r.Detail)
		}
	}
	return allOK
}
