package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestCheckLink(t *testing.T) {
	dir := t.TempDir()

	realTarget := filepath.Join(dir, "real")
	if err := os.WriteFile(realTarget, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	okPath := filepath.Join(dir, "ok")
	if err := os.Symlink(realTarget, okPath); err != nil {
		t.Fatal(err)
	}

	wrongPath := filepath.Join(dir, "wrong")
	otherTarget := filepath.Join(dir, "other")
	if err := os.WriteFile(otherTarget, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(otherTarget, wrongPath); err != nil {
		t.Fatal(err)
	}

	notSymlinkPath := filepath.Join(dir, "plain")
	if err := os.WriteFile(notSymlinkPath, []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	brokenPath := filepath.Join(dir, "broken")
	if err := os.Symlink(filepath.Join(dir, "does-not-exist"), brokenPath); err != nil {
		t.Fatal(err)
	}

	missingPath := filepath.Join(dir, "missing")

	relPath := filepath.Join(dir, "rel")
	if err := os.Symlink("real", relPath); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		link Link
		want string
	}{
		{"ok", Link{Path: okPath, Target: realTarget}, "ok"},
		{"relative target resolved against link dir", Link{Path: relPath, Target: "real"}, "ok"},
		{"missing", Link{Path: missingPath, Target: realTarget}, "missing"},
		{"not a symlink", Link{Path: notSymlinkPath, Target: realTarget}, "not-symlink"},
		{"wrong target", Link{Path: wrongPath, Target: realTarget}, "wrong-target"},
		{"broken", Link{Path: brokenPath, Target: filepath.Join(dir, "does-not-exist")}, "broken"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkLink(tt.link)
			if got.Status != tt.want {
				t.Errorf("checkLink(%+v) = %q, want %q", tt.link, got.Status, tt.want)
			}
		})
	}
}

func TestWriteCheckReport(t *testing.T) {
	results := []checkResult{
		{Link: Link{Path: "/a", Target: "/x"}, Status: "ok"},
		{Link: Link{Path: "/b", Target: "/y"}, Status: "missing"},
	}

	var buf bytes.Buffer
	ok := writeCheckReport(&buf, results)
	if ok {
		t.Error("writeCheckReport() = true, want false when a link is missing")
	}
	if !bytes.Contains(buf.Bytes(), []byte("/a")) || !bytes.Contains(buf.Bytes(), []byte("/b")) {
		t.Errorf("report missing expected paths: %s", buf.String())
	}

	buf.Reset()
	ok = writeCheckReport(&buf, results[:1])
	if !ok {
		t.Error("writeCheckReport() = false, want true when every link is ok")
	}
}
