package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// parseStowTree reads the top-level entries of a GNU Stow-style package
// directory and turns each into a Link from targetDir into the package.
//
// This mirrors stow's own behavior against a clean target: as long as
// nothing already exists at targetDir/name, stow links the whole subtree
// there in one symlink instead of descending into it (tree folding). It's
// how the README's own nvim example works - one link for the whole
// directory rather than one per file inside it.
//
// What this doesn't model is unfolding: real stow will descend into a
// package subdirectory and link individual files if targetDir/name is
// already a real directory rather than a symlink, because in that case
// the single-symlink shortcut would clobber whatever's already there.
// That decision depends on live state of the target, and this function
// only reads the package - it doesn't look at the target at all.
func parseStowTree(pkgDir, targetDir string) ([]Link, error) {
	entries, err := os.ReadDir(pkgDir)
	if err != nil {
		return nil, fmt.Errorf("reading stow package %s: %w", pkgDir, err)
	}
	var links []Link
	for _, e := range entries {
		name := e.Name()
		if name == ".stow-local-ignore" {
			continue
		}
		links = append(links, Link{
			Path:   filepath.Join(targetDir, name),
			Target: filepath.Join(pkgDir, name),
		})
	}
	return links, nil
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
