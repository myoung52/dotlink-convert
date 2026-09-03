package main

import (
	"os"
	"strings"
)

// expandPath expands a leading ~ or $HOME in p to the user's home
// directory, the same way an interactive shell would expand them in
// an unquoted argument. Anything not at the start of the string is
// left alone - this isn't a general variable expander, just enough
// to undo the two forms people actually write in a manifest.
func expandPath(p string) string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return p
	}
	switch {
	case p == "~":
		return home
	case strings.HasPrefix(p, "~/"):
		return home + p[1:]
	case p == "$HOME":
		return home
	case strings.HasPrefix(p, "$HOME/"):
		return home + p[len("$HOME"):]
	default:
		return p
	}
}
