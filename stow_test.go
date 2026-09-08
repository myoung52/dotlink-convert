package main

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestParseStowTree(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(pkgDir, ".vimrc"), []byte("vim config"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(pkgDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".stow-local-ignore"), []byte("*.swp"), 0o644); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	sort.Slice(links, func(i, j int) bool { return links[i].Path < links[j].Path })

	want := []Link{
		{Path: filepath.Join(targetDir, ".config"), Target: filepath.Join(pkgDir, ".config")},
		{Path: filepath.Join(targetDir, ".vimrc"), Target: filepath.Join(pkgDir, ".vimrc")},
	}
	if len(links) != len(want) {
		t.Fatalf("parseStowTree() = %+v, want %+v", links, want)
	}
	for i := range want {
		if links[i] != want[i] {
			t.Errorf("link %d = %+v, want %+v", i, links[i], want[i])
		}
	}
}

func TestParseStowTreeMissingPackage(t *testing.T) {
	_, err := parseStowTree(filepath.Join(t.TempDir(), "does-not-exist"), t.TempDir())
	if err == nil {
		t.Fatal("parseStowTree() with missing package dir = nil error, want error")
	}
}

func TestReadStowInputRequiresPackageDir(t *testing.T) {
	_, err := readStowInput("", "")
	if err == nil {
		t.Fatal("readStowInput(\"\", \"\") = nil error, want error")
	}
}

func TestReadStowInputExplicitTarget(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(pkgDir, ".zshrc"), []byte("zsh config"), 0o644); err != nil {
		t.Fatal(err)
	}

	links, err := readStowInput(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("readStowInput: %v", err)
	}
	want := Link{Path: filepath.Join(targetDir, ".zshrc"), Target: filepath.Join(pkgDir, ".zshrc")}
	if len(links) != 1 || links[0] != want {
		t.Errorf("readStowInput() = %+v, want [%+v]", links, want)
	}
}
