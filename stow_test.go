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

func TestParseStowTreeUnfoldsExistingRealDir(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	if err := os.Mkdir(filepath.Join(pkgDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".config", "nvim.conf"), []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".zshrc"), []byte("zsh"), 0o644); err != nil {
		t.Fatal(err)
	}

	// targetDir/.config already exists as a real directory (e.g. from
	// another package), so stow must descend into it instead of folding.
	if err := os.Mkdir(filepath.Join(targetDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	sort.Slice(links, func(i, j int) bool { return links[i].Path < links[j].Path })

	want := []Link{
		{Path: filepath.Join(targetDir, ".config", "nvim.conf"), Target: filepath.Join(pkgDir, ".config", "nvim.conf")},
		{Path: filepath.Join(targetDir, ".zshrc"), Target: filepath.Join(pkgDir, ".zshrc")},
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

func TestParseStowTreeFoldsWhenTargetIsSymlink(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()
	otherPkg := t.TempDir()

	if err := os.Mkdir(filepath.Join(pkgDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".config", "nvim.conf"), []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}

	// targetDir/.config is already a symlink (e.g. folded from a prior
	// stow run), not a real directory, so it should still fold rather
	// than unfold.
	if err := os.Symlink(otherPkg, filepath.Join(targetDir, ".config")); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	want := Link{Path: filepath.Join(targetDir, ".config"), Target: filepath.Join(pkgDir, ".config")}
	if len(links) != 1 || links[0] != want {
		t.Errorf("parseStowTree() = %+v, want [%+v]", links, want)
	}
}

func TestParseStowTreeUnfoldsNested(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	nested := filepath.Join(pkgDir, ".config", "nvim")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "init.lua"), []byte("-- init"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Both targetDir/.config and targetDir/.config/nvim already exist as
	// real directories, so unfolding must recurse two levels deep.
	if err := os.MkdirAll(filepath.Join(targetDir, ".config", "nvim"), 0o755); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	want := Link{
		Path:   filepath.Join(targetDir, ".config", "nvim", "init.lua"),
		Target: filepath.Join(pkgDir, ".config", "nvim", "init.lua"),
	}
	if len(links) != 1 || links[0] != want {
		t.Errorf("parseStowTree() = %+v, want [%+v]", links, want)
	}
}

func TestParseStowTreeIgnoresGlobPatterns(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(pkgDir, ".stow-local-ignore"), []byte("*.swp\n# comment\n.git\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".vimrc"), []byte("vim config"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, ".vimrc.swp"), []byte("swap"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(pkgDir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	want := Link{Path: filepath.Join(targetDir, ".vimrc"), Target: filepath.Join(pkgDir, ".vimrc")}
	if len(links) != 1 || links[0] != want {
		t.Errorf("parseStowTree() = %+v, want [%+v]", links, want)
	}
}

func TestParseStowTreeIgnorePatternAppliesWhenUnfolding(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(pkgDir, ".stow-local-ignore"), []byte("*.orig\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(pkgDir, ".config")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "init.lua"), []byte("-- init"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "init.lua.orig"), []byte("-- backup"), 0o644); err != nil {
		t.Fatal(err)
	}

	// targetDir/.config already exists as a real directory, forcing
	// unfold, so the top-level ignore pattern must still apply once
	// parseStowTree recurses into it.
	if err := os.Mkdir(filepath.Join(targetDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}

	links, err := parseStowTree(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("parseStowTree: %v", err)
	}

	want := Link{
		Path:   filepath.Join(targetDir, ".config", "init.lua"),
		Target: filepath.Join(pkgDir, ".config", "init.lua"),
	}
	if len(links) != 1 || links[0] != want {
		t.Errorf("parseStowTree() = %+v, want [%+v]", links, want)
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

func TestStowReport(t *testing.T) {
	pkgDir := t.TempDir()
	targetDir := t.TempDir()

	// .vimrc has nothing at targetDir/.vimrc, so it folds. .config
	// already exists as a real directory at the target, so it must
	// unfold, and its own children are reported in turn.
	if err := os.WriteFile(filepath.Join(pkgDir, ".vimrc"), []byte("vim config"), 0o644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(pkgDir, ".config")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "nvim.conf"), []byte("config"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(targetDir, ".config"), 0o755); err != nil {
		t.Fatal(err)
	}

	report, err := stowReport(pkgDir, targetDir)
	if err != nil {
		t.Fatalf("stowReport: %v", err)
	}

	sort.Slice(report, func(i, j int) bool { return report[i].Path < report[j].Path })

	want := []stowReportEntry{
		{Path: filepath.Join(targetDir, ".config"), Action: "unfold"},
		{Path: filepath.Join(targetDir, ".config", "nvim.conf"), Action: "fold"},
		{Path: filepath.Join(targetDir, ".vimrc"), Action: "fold"},
	}
	if len(report) != len(want) {
		t.Fatalf("stowReport() = %+v, want %+v", report, want)
	}
	for i := range want {
		if report[i] != want[i] {
			t.Errorf("entry %d = %+v, want %+v", i, report[i], want[i])
		}
	}
}

func TestStowReportRequiresPackageDir(t *testing.T) {
	_, err := stowReport("", "")
	if err == nil {
		t.Fatal("stowReport(\"\", \"\") = nil error, want error")
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
