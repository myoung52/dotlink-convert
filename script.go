package main

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// parseScript reads a shell script and pulls a Link out of every line
// that invokes ln. Flags (anything starting with -) are ignored; the
// two remaining path arguments are taken as target and link path, in
// that order, matching how ln itself takes them.
func parseScript(r io.Reader) ([]Link, error) {
	var links []Link
	scanner := bufio.NewScanner(r)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		words := splitShellWords(line)
		if len(words) == 0 || words[0] != "ln" {
			continue
		}
		var args []string
		for _, w := range words[1:] {
			if strings.HasPrefix(w, "-") {
				continue
			}
			args = append(args, w)
		}
		if len(args) != 2 {
			return nil, fmt.Errorf("script line %d: expected ln with a target and a link path, got %d path argument(s)", lineNum, len(args))
		}
		links = append(links, Link{Target: args[0], Path: args[1]})
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return links, nil
}

func writeScript(w io.Writer, links []Link) error {
	if _, err := fmt.Fprintln(w, "#!/bin/sh"); err != nil {
		return err
	}
	for _, l := range links {
		if _, err := fmt.Fprintf(w, "ln -sf %s %s\n", shellQuote(l.Target), shellQuote(l.Path)); err != nil {
			return err
		}
	}
	return nil
}

// shellQuote wraps a path in double quotes if it contains anything a
// plain whitespace split would misread.
func shellQuote(s string) string {
	if s == "" || strings.ContainsAny(s, " \t\"'$`\\") {
		r := strings.NewReplacer(`\`, `\\`, `"`, `\"`, `$`, `\$`, "`", "\\`")
		return `"` + r.Replace(s) + `"`
	}
	return s
}

// splitShellWords does a minimal word split that understands single
// and double quotes. It's enough for the "ln -sf a b" lines this tool
// needs to read, not a general-purpose shell parser.
func splitShellWords(line string) []string {
	var words []string
	var cur strings.Builder
	inWord := false
	var quote rune

	flush := func() {
		if inWord {
			words = append(words, cur.String())
			cur.Reset()
			inWord = false
		}
	}

	for _, c := range line {
		if quote != 0 {
			if c == quote {
				quote = 0
				continue
			}
			cur.WriteRune(c)
			continue
		}
		switch {
		case c == '\'' || c == '"':
			quote = c
			inWord = true
		case c == ' ' || c == '\t':
			flush()
		default:
			inWord = true
			cur.WriteRune(c)
		}
	}
	flush()
	return words
}
