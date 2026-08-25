package main

import (
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
	var path string
	if len(os.Args) >= 3 {
		path = os.Args[2]
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
			convErr = writeScript(os.Stdout, links)
		}
	case "to-manifest":
		links, convErr = parseScript(in)
		if convErr == nil {
			convErr = writeManifest(os.Stdout, links)
		}
	default:
		usage()
		os.Exit(2)
	}

	if convErr != nil {
		fmt.Fprintln(os.Stderr, "dotlink:", convErr)
		os.Exit(1)
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
	fmt.Fprintln(os.Stderr, "usage: dotlink to-script|to-manifest [file]")
	fmt.Fprintln(os.Stderr, "  reads from file, or stdin if omitted or file is -")
}
