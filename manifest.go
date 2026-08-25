package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// Link is one symlink entry. Path is where the symlink is created,
// Target is the file or directory it points to (ln's argument order
// is target-then-link, so Target comes first there too).
type Link struct {
	Path   string
	Target string
}

// parseManifest reads lines of the form "path = target". Blank lines
// and lines starting with # are skipped.
func parseManifest(r io.Reader) ([]Link, error) {
	var links []Link
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		idx := strings.Index(line, "=")
		if idx < 0 {
			return nil, fmt.Errorf("manifest line %d: missing '=': %q", lineNum, line)
		}
		path := strings.TrimSpace(line[:idx])
		target := strings.TrimSpace(line[idx+1:])
		if path == "" || target == "" {
			return nil, fmt.Errorf("manifest line %d: empty path or target", lineNum)
		}
		links = append(links, Link{Path: path, Target: target})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func writeManifest(w io.Writer, links []Link) error {
	for _, l := range links {
		if _, err := fmt.Fprintf(w, "%s = %s\n", l.Path, l.Target); err != nil {
			return err
		}
	}
	return nil
}
