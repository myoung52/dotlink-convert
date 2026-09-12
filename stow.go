package main

import (
	"fmt"
	"os"
	"path/filepath"
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
		pkgPath := filepath.Join(pkgDir, name)
		targetPath := filepath.Join(targetDir, name)

		if e.IsDir() && targetIsRealDir(targetPath) {
			sub, err := parseStowTree(pkgPath, targetPath)
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
