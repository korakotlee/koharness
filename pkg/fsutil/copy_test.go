package fsutil

import (
	"testing"

	"github.com/spf13/afero"
)

func TestCopyFile(t *testing.T) {
	fs := afero.NewMemMapFs()
	src := "/source/file.txt"
	dst := "/target/file.txt"
	content := []byte("hello world")

	if err := afero.WriteFile(fs, src, content, 0644); err != nil {
		t.Fatalf("failed to write source file: %v", err)
	}

	if err := CopyFile(fs, src, dst); err != nil {
		t.Fatalf("CopyFile failed: %v", err)
	}

	got, err := afero.ReadFile(fs, dst)
	if err != nil {
		t.Fatalf("failed to read target file: %v", err)
	}

	if string(got) != string(content) {
		t.Errorf("content mismatch: got %q, want %q", string(got), string(content))
	}
}

func TestCopyDir(t *testing.T) {
	fs := afero.NewMemMapFs()
	srcDir := "/source/dir"
	dstDir := "/target/dir"

	if err := afero.WriteFile(fs, srcDir+"/file1.txt", []byte("file1"), 0644); err != nil {
		t.Fatalf("failed to write file1: %v", err)
	}
	if err := afero.WriteFile(fs, srcDir+"/sub/file2.txt", []byte("file2"), 0644); err != nil {
		t.Fatalf("failed to write file2: %v", err)
	}

	if err := CopyDir(fs, srcDir, dstDir); err != nil {
		t.Fatalf("CopyDir failed: %v", err)
	}

	got1, err := afero.ReadFile(fs, dstDir+"/file1.txt")
	if err != nil || string(got1) != "file1" {
		t.Errorf("file1 copy failed: %v, content: %s", err, string(got1))
	}

	got2, err := afero.ReadFile(fs, dstDir+"/sub/file2.txt")
	if err != nil || string(got2) != "file2" {
		t.Errorf("file2 copy failed: %v, content: %s", err, string(got2))
	}
}
