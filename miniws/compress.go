package miniws

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type compress struct{}

// takes a file path, compresses the file into .tar.gz format
// creating the archive in the same directory with the same base name.
func (c *compress) CompressFile(filePath string) error {
	srcFile, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening source file: %w", err)
	}
	defer srcFile.Close()

	// destination file path: same dir, same base name with .tar.gz
	dir := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	destName := base + ".tar.gz"
	destPath := filepath.Join(dir, destName)

	destFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("creating dest file: %w", err)
	}

	defer func() {
		if cerr := destFile.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing dest file: %w", cerr)
		}
	}()

	// gzip writer
	gzWriter := gzip.NewWriter(destFile)
	defer func() {
		if cerr := gzWriter.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing gzip writer: %w", cerr)
		}
	}()

	// tar writer
	tarWriter := tar.NewWriter(gzWriter)
	defer func() {
		if cerr := tarWriter.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("closing tar writer: %w", cerr)
		}
	}()

	// Get file info for header
	fileInfo, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("getting file info: %w", err)
	}

	// Prepare tar header
	head := &tar.Header{
		Name:    base,
		Size:    fileInfo.Size(),
		Mode:    int64(fileInfo.Mode().Perm()),
		ModTime: fileInfo.ModTime(),
	}
	if err := tarWriter.WriteHeader(head); err != nil {
		return fmt.Errorf("writing tar header: %w", err)
	}

	// Copy file data into tar
	if _, err := io.Copy(tarWriter, srcFile); err != nil {
		return fmt.Errorf("writing file data to tar: %w", err)
	}

	return nil
}
