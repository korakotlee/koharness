// Package fsutil provides unified file system utilities for copying files and directory structures across afero filesystems.
package fsutil

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
)

// CopyFile copies a single file from src to dst on the specified afero.Fs filesystem, preserving file mode.
func CopyFile(fs afero.Fs, src, dst string) error {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	if err := fs.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("failed to create destination parent directory for %s: %w", dst, err)
	}

	srcFile, err := fs.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file %s: %w", src, err)
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("failed to stat source file %s: %w", src, err)
	}

	dstFile, err := fs.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file %s: %w", dst, err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed copying content from %s to %s: %w", src, dst, err)
	}

	return nil
}

// CopyDir recursively copies a directory and its contents from src to dst on the specified afero.Fs filesystem.
func CopyDir(fs afero.Fs, src, dst string) error {
	if fs == nil {
		fs = afero.NewOsFs()
	}
	if err := fs.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory %s: %w", dst, err)
	}

	entries, err := afero.ReadDir(fs, src)
	if err != nil {
		return fmt.Errorf("failed to read source directory %s: %w", src, err)
	}

	for _, entry := range entries {
		srcChild := filepath.Join(src, entry.Name())
		dstChild := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := CopyDir(fs, srcChild, dstChild); err != nil {
				return err
			}
		} else {
			if err := CopyFile(fs, srcChild, dstChild); err != nil {
				return err
			}
		}
	}
	return nil
}
