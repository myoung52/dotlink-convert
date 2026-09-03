package main

import "testing"

func TestExpandPath(t *testing.T) {
	t.Setenv("HOME", "/home/test")

	cases := []struct {
		in   string
		want string
	}{
		{"~", "/home/test"},
		{"~/.vimrc", "/home/test/.vimrc"},
		{"$HOME", "/home/test"},
		{"$HOME/.vimrc", "/home/test/.vimrc"},
		{"dotfiles/vim/vimrc", "dotfiles/vim/vimrc"},
		{"~notme/.vimrc", "~notme/.vimrc"},
		{"a/~/b", "a/~/b"},
	}
	for _, c := range cases {
		if got := expandPath(c.in); got != c.want {
			t.Errorf("expandPath(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}
