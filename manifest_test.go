package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseManifest(t *testing.T) {
	input := `# ~/.dotlinks
~/.vimrc = dotfiles/vim/vimrc
~/.zshrc=dotfiles/zsh/zshrc

# comment in the middle
~/.config/nvim = dotfiles/nvim
`
	links, err := parseManifest(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}
	want := []Link{
		{Path: "~/.vimrc", Target: "dotfiles/vim/vimrc"},
		{Path: "~/.zshrc", Target: "dotfiles/zsh/zshrc"},
		{Path: "~/.config/nvim", Target: "dotfiles/nvim"},
	}
	if len(links) != len(want) {
		t.Fatalf("got %d links, want %d: %+v", len(links), len(want), links)
	}
	for i := range want {
		if links[i] != want[i] {
			t.Errorf("link %d: got %+v, want %+v", i, links[i], want[i])
		}
	}
}

func TestParseManifestMissingEquals(t *testing.T) {
	_, err := parseManifest(strings.NewReader("~/.vimrc dotfiles/vim/vimrc\n"))
	if err == nil {
		t.Fatal("expected error for line missing '='")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("error should reference line number, got: %v", err)
	}
}

func TestParseManifestEmptyPathOrTarget(t *testing.T) {
	cases := []string{
		" = dotfiles/vim/vimrc\n",
		"~/.vimrc = \n",
		"=\n",
	}
	for _, c := range cases {
		if _, err := parseManifest(strings.NewReader(c)); err == nil {
			t.Errorf("expected error for input %q", c)
		}
	}
}

func TestParseManifestValueContainingEquals(t *testing.T) {
	// only the first '=' should be treated as the separator
	links, err := parseManifest(strings.NewReader("~/.env = dotfiles/env?key=value\n"))
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].Target != "dotfiles/env?key=value" {
		t.Errorf("target = %q, want %q", links[0].Target, "dotfiles/env?key=value")
	}
}

func TestWriteManifest(t *testing.T) {
	links := []Link{
		{Path: "~/.vimrc", Target: "dotfiles/vim/vimrc"},
		{Path: "~/.zshrc", Target: "dotfiles/zsh/zshrc"},
	}
	var buf bytes.Buffer
	if err := writeManifest(&buf, links); err != nil {
		t.Fatalf("writeManifest: %v", err)
	}
	want := "~/.vimrc = dotfiles/vim/vimrc\n~/.zshrc = dotfiles/zsh/zshrc\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestManifestRoundTrip(t *testing.T) {
	links := []Link{
		{Path: "~/.vimrc", Target: "dotfiles/vim/vimrc"},
		{Path: "~/.config/nvim", Target: "dotfiles/nvim"},
	}
	var buf bytes.Buffer
	if err := writeManifest(&buf, links); err != nil {
		t.Fatalf("writeManifest: %v", err)
	}
	got, err := parseManifest(&buf)
	if err != nil {
		t.Fatalf("parseManifest: %v", err)
	}
	if len(got) != len(links) {
		t.Fatalf("got %d links, want %d", len(got), len(links))
	}
	for i := range links {
		if got[i] != links[i] {
			t.Errorf("link %d: got %+v, want %+v", i, got[i], links[i])
		}
	}
}
