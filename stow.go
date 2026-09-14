package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// parseStowTree reads a GNU Stow-style package directory and turns it
// into Links from targetDir into the package, mirroring stow's own tree
// folding and unfolding against a live target.
//
// As long as nothing already exists at targetDir/name, stow links the
// whole subtree there in one symlink instead of descending into it (tree
// folding). It's how the README's own nvim example works - one link for
// the whole directory rather than one per file inside it.
//
// But if targetDir/name is already a real directory rather than a
// symlink, folding it would clobber whatever's already there, so stow
// unfolds instead: it descends into the package subdirectory and links
// individual entries inside targetDir/name, recursing again at each
// level. This is why parseStowTree takes a live targetDir rather than
// just reading the package - the decision depends on what's actually on
// disk.
func parseStowTree(pkgDir, targetDir string) ([]Link, error) {
	patterns, err := loadIgnorePatterns(pkgDir)
	if err != nil {
		return nil, err
	}
	return parseStowTreeIgnoring(pkgDir, targetDir, patterns)
}

// parseStowTreeIgnoring does the actual tree walk, skipping any entry
// whose basename matches one of patterns. patterns come from the
// package's top-level .stow-local-ignore and are reused unchanged at
// every recursion depth, matching how GNU Stow itself applies one
// ignore list against basenames throughout the whole package rather
// than re-reading it per directory.
func parseStowTreeIgnoring(pkgDir, targetDir string, patterns []string) ([]Link, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("reading stow package %s: %w", pkgDir, err)
	}
	var links []Link
	for _, e := range entries {
		name := e.Name()
		if name == ".stow-local-ignore" || matchesIgnore(name, patterns) {
			continue
		}
		pkgPath := filepath.Join(pkgDir, name)
		targetPath := filepath.Join(targetDir, name)

		if e.IsDir() && targetIsRealDir(targetPath) {
			sub, err := parseStowTreeIgnoring(pkgPath, targetPath, patterns)
			if err != nil {
				return nil, err
			}
			links = append(links, sub...)
			continue
		}

		links = append(links, Link{Path: targetPath, Target: pkgPath})
	}
	return links, nil
}

// loadIgnorePatterns reads glob patterns from pkgDir/.stow-local-ignore,
// one per line, with blank lines and lines starting with # skipped. Real
// GNU Stow ignore files hold Perl regexes; matching those without
// pulling in a regex engine isn't worth it here, so patterns are
// filepath.Match globs instead - enough to cover the common cases like
// "*.swp" or ".git".
func loadIgnorePatterns(pkgDir string) ([]string, error) {
	data, err := os.ReadFile(filepath.Join(pkgDir, ".stow-local-ignore"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reading .stow-local-ignore: %w", err)
	}
	var patterns []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		patterns = append(patterns, line)
	}
	return patterns, nil
}

// matchesIgnore reports whether name matches any of patterns. A
// malformed pattern is treated as a non-match rather than an error,
// same as filepath.Match's own ErrBadPattern handling elsewhere in the
// stdlib.
func matchesIgnore(name string, patterns []string) bool {
	for _, p := range patterns {
		if ok, err := filepath.Match(p, name); err == nil && ok {
			return true
		}
	}
	return false
}

// targetIsRealDir reports whether targetPath already exists on disk as a
// directory that isn't itself a symlink - the condition under which stow
// must unfold rather than fold.
func targetIsRealDir(targetPath string) bool {
	info, err := os.Lstat(targetPath)
	if err != nil {
		return false
	}
	return info.IsDir() && info.Mode()&os.ModeSymlink == 0
}

// readStowInput resolves the target directory for a stow package and
// reads it. An explicit -target always wins; otherwise it falls back to
// the user's home directory, matching stow's own default of linking into
// the parent of the package directory's parent.
func readStowInput(pkgDir, target string) ([]Link, error) {
	if pkgDir == "" {
		return nil, fmt.Errorf("stow format requires a package directory argument, not stdin")
	}
	if target == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("resolving default -target: %w", err)
		}
		target = home
	}
	return parseStowTree(pkgDir, target)
}
