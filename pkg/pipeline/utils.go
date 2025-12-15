package pipeline

import (
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
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

// StartSpinner renders a simple CLI spinner with a given prefix message.
// It returns a stop function. Call stop with a final status string (e.g., "done" or "failed").
// The spinner runs in a goroutine and will be cleaned up synchronously when stop is called.
func StartSpinner(prefix string) func(final string) {
	frames := []rune{'|', '/', '-', '\\'}
	ticker := time.NewTicker(120 * time.Millisecond)
	done := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-ticker.C:
				fmt.Printf("\r%s %c", prefix, frames[i%len(frames)])
				i++
			case <-done:
				ticker.Stop()
				return
			}
		}
	}()

	return func(final string) {
		close(done)
		wg.Wait()
		// Clear spinner line and print final status
		fmt.Printf("\r%s %s\n", prefix, final)
	}
}

// splitArgs is a tiny helper that splits a string on whitespace.
// It is used to expand custom Trimmomatic parameters supplied by the user.
func splitArgs(s string) []string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil
	}
	return fields
}
