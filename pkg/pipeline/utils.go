package pipeline

import (
	"bufio"
	"compress/gzip"
	"io"
	"os"
)

// fileExists checks if a file exists and is not a directory.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// gzipIntegrityOK verifies that a .gz file can be fully decompressed without error.
// It reads the entire stream and discards the output to ensure the trailer is valid
// (detects truncated/corrupted gzip files).
func gzipIntegrityOK(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	// Quick zero-length check
	if stat, err := f.Stat(); err == nil {
		if stat.Size() == 0 {
			return false
		}
	}

	gr, err := gzip.NewReader(f)
	if err != nil {
		return false
	}
	defer gr.Close()

	// Use a buffered copy to /dev/null (io.Discard) to force full stream read
	bufReader := bufio.NewReader(gr)
	if _, err := io.Copy(io.Discard, bufReader); err != nil {
		return false
	}
	return true
}

// removeIfExists deletes the file if it exists (ignores directories).
func removeIfExists(path string) error {
	if fileExists(path) {
		return os.Remove(path)
	}
	return nil
}
