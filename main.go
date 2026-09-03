package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	cmd := os.Args[1]
	if cmd != "to-script" && cmd != "to-manifest" {
		usage()
		os.Exit(2)
	}

	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	expand := fs.Bool("expand", false, "expand leading ~ and $HOME in paths and targets")
	fs.Usage = usage
	fs.Parse(os.Args[2:])

	var path string
	if fs.NArg() > 0 {
		path = fs.Arg(0)
	}

	in, err := openInput(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "dotlink:", err)
		os.Exit(1)
	}
	defer in.Close()

	var links []Link
	var convErr error

	switch cmd {
	case "to-script":
		links, convErr = parseManifest(in)
		if convErr == nil {
			if *expand {
				expandLinks(links)
			}
			convErr = writeScript(os.Stdout, links)
		}
	case "to-manifest":
		links, convErr = parseScript(in)
		if convErr == nil {
			if *expand {
				expandLinks(links)
			}
			convErr = writeManifest(os.Stdout, links)
		}
	}

	if convErr != nil {
		fmt.Fprintln(os.Stderr, "dotlink:", convErr)
		os.Exit(1)
	}
}

// expandLinks rewrites every path and target in place using expandPath.
func expandLinks(links []Link) {
	for i := range links {
		links[i].Path = expandPath(links[i].Path)
		links[i].Target = expandPath(links[i].Target)
	}
}

// openInput reads from the given path, or from stdin when path is
// empty or "-". That's the one thing this tool needs to get right:
// every pipeline that already produces one of these formats should
// be able to feed it straight in without a temp file.
func openInput(path string) (io.ReadCloser, error) {
	if path == "" || path == "-" {
		return io.NopCloser(os.Stdin), nil
	}
	return os.Open(path)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: dotlink to-script|to-manifest [-expand] [file]")
	fmt.Fprintln(os.Stderr, "  reads from file, or stdin if omitted or file is -")
	fmt.Fprintln(os.Stderr, "  -expand   expand leading ~ and $HOME in paths and targets")
}
