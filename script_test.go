package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestParseScript(t *testing.T) {
	input := `#!/bin/sh
# set up dotfile links
ln -sf dotfiles/vim/vimrc ~/.vimrc
ln -sf dotfiles/zsh/zshrc ~/.zshrc

ln -s dotfiles/nvim ~/.config/nvim
`
	links, err := parseScript(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseScript: %v", err)
	}
	want := []Link{
		{Target: "dotfiles/vim/vimrc", Path: "~/.vimrc"},
		{Target: "dotfiles/zsh/zshrc", Path: "~/.zshrc"},
		{Target: "dotfiles/nvim", Path: "~/.config/nvim"},
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

func TestParseScriptIgnoresNonLnLines(t *testing.T) {
	input := `#!/bin/sh
mkdir -p ~/.config
ln -sf dotfiles/nvim ~/.config/nvim
echo done
`
	links, err := parseScript(strings.NewReader(input))
	if err != nil {
		t.Fatalf("parseScript: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1: %+v", len(links), links)
	}
}

func TestParseScriptWrongArgCount(t *testing.T) {
	cases := []string{
		"ln -sf onlyonepath\n",
		"ln -sf a b c\n",
		"ln\n",
	}
	for _, c := range cases {
		if _, err := parseScript(strings.NewReader(c)); err == nil {
			t.Errorf("expected error for input %q", c)
		}
	}
}

func TestParseScriptQuotedPaths(t *testing.T) {
	links, err := parseScript(strings.NewReader(`ln -sf "dotfiles/my app" '~/.my app'` + "\n"))
	if err != nil {
		t.Fatalf("parseScript: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("got %d links, want 1", len(links))
	}
	if links[0].Target != "dotfiles/my app" || links[0].Path != "~/.my app" {
		t.Errorf("got %+v", links[0])
	}
}

func TestWriteScript(t *testing.T) {
	links := []Link{
		{Target: "dotfiles/vim/vimrc", Path: "~/.vimrc"},
		{Target: "dotfiles/my app", Path: "~/.my app"},
	}
	var buf bytes.Buffer
	if err := writeScript(&buf, links); err != nil {
		t.Fatalf("writeScript: %v", err)
	}
	want := "#!/bin/sh\n" +
		"ln -sf dotfiles/vim/vimrc ~/.vimrc\n" +
		`ln -sf "dotfiles/my app" "~/.my app"` + "\n"
	if buf.String() != want {
		t.Errorf("got %q, want %q", buf.String(), want)
	}
}

func TestScriptRoundTrip(t *testing.T) {
	links := []Link{
		{Target: "dotfiles/vim/vimrc", Path: "~/.vimrc"},
		{Target: "dotfiles/my app", Path: "~/.my app"},
	}
	var buf bytes.Buffer
	if err := writeScript(&buf, links); err != nil {
		t.Fatalf("writeScript: %v", err)
	}
	got, err := parseScript(&buf)
	if err != nil {
		t.Fatalf("parseScript: %v", err)
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

func TestSplitShellWords(t *testing.T) {
	cases := []struct {
		in   string
		want []string
	}{
		{`ln -sf a b`, []string{"ln", "-sf", "a", "b"}},
		{`ln -sf "a b" c`, []string{"ln", "-sf", "a b", "c"}},
		{`ln -sf 'a b' "c d"`, []string{"ln", "-sf", "a b", "c d"}},
		{``, nil},
		{`   `, nil},
	}
	for _, c := range cases {
		got := splitShellWords(c.in)
		if len(got) != len(c.want) {
			t.Errorf("splitShellWords(%q) = %v, want %v", c.in, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("splitShellWords(%q)[%d] = %q, want %q", c.in, i, got[i], c.want[i])
			}
		}
	}
}
