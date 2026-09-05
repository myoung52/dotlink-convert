package main

import (
	"strings"
	"testing"
)

func TestDuplicatePathsNone(t *testing.T) {
	links := []Link{
		{Path: "~/.vimrc", Target: "dotfiles/vim/vimrc"},
		{Path: "~/.zshrc", Target: "dotfiles/zsh/zshrc"},
	}
	if err := duplicatePaths(links); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDuplicatePathsFound(t *testing.T) {
	links := []Link{
		{Path: "~/.vimrc", Target: "dotfiles/vim/vimrc"},
		{Path: "~/.zshrc", Target: "dotfiles/zsh/zshrc"},
		{Path: "~/.vimrc", Target: "dotfiles/other/vimrc"},
	}
	err := duplicatePaths(links)
	if err == nil {
		t.Fatal("expected an error for duplicate path")
	}
	if want := "~/.vimrc"; !strings.Contains(err.Error(), want) {
		t.Errorf("error %q does not mention %q", err.Error(), want)
	}
}

func TestDuplicatePathsReportsEachOnlyOnce(t *testing.T) {
	// three entries sharing a path should still name it once, not
	// once per repeat.
	links := []Link{
		{Path: "~/.vimrc", Target: "a"},
		{Path: "~/.vimrc", Target: "b"},
		{Path: "~/.vimrc", Target: "c"},
	}
	err := duplicatePaths(links)
	if err == nil {
		t.Fatal("expected an error for duplicate path")
	}
	if count := strings.Count(err.Error(), "~/.vimrc"); count != 1 {
		t.Errorf("expected path named once, got %d times in %q", count, err.Error())
	}
}

func TestDuplicatePathsMultiple(t *testing.T) {
	links := []Link{
		{Path: "~/.vimrc", Target: "a"},
		{Path: "~/.vimrc", Target: "b"},
		{Path: "~/.zshrc", Target: "c"},
		{Path: "~/.zshrc", Target: "d"},
	}
	err := duplicatePaths(links)
	if err == nil {
		t.Fatal("expected an error for duplicate paths")
	}
	for _, want := range []string{"~/.vimrc", "~/.zshrc"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not mention %q", err.Error(), want)
		}
	}
}
