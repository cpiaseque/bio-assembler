package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type TrimmomaticStep struct {
	InputFq1         string
	InputFq2         string
	PairedOutput1    string
	PairedOutput2    string
	UnpairedOutput1  string
	UnpairedOutput2  string
	Threads          int
	AdapterFastaPath string
}

func (s *TrimmomaticStep) Name() string {
	return "Trimmomatic"
}

func (s *TrimmomaticStep) Run() error {
	if !fileExists(s.InputFq1) || !fileExists(s.InputFq2) {
		return fmt.Errorf("input FASTQ files not found: %s, %s", s.InputFq1, s.InputFq2)
	}
	if s.AdapterFastaPath == "" || !fileExists(s.AdapterFastaPath) {
		return fmt.Errorf("adapter FASTA not found: %s", s.AdapterFastaPath)
	}
	if fileExists(s.PairedOutput1) && fileExists(s.PairedOutput2) {
		// Validate gzip integrity to avoid using truncated outputs from a previous failed run
		if gzipIntegrityOK(s.PairedOutput1) && gzipIntegrityOK(s.PairedOutput2) {
			fmt.Println("Trimmed files already exist and passed integrity check, skipping Trimmomatic.")
			return nil
		}
		fmt.Println("Existing trimmed files appear corrupted or unfinished; re-generating with Trimmomatic...")
		// Best effort cleanup of previous outputs
		_ = removeIfExists(s.PairedOutput1)
		_ = removeIfExists(s.PairedOutput2)
		_ = removeIfExists(s.UnpairedOutput1)
		_ = removeIfExists(s.UnpairedOutput2)
	}

	fmt.Println("Running Trimmomatic for read trimming...")

	// Ensure output directories exist
	outDirs := map[string]struct{}{
		filepath.Dir(s.PairedOutput1):   {},
		filepath.Dir(s.PairedOutput2):   {},
		filepath.Dir(s.UnpairedOutput1): {},
		filepath.Dir(s.UnpairedOutput2): {},
	}
	for d := range outDirs {
		if err := os.MkdirAll(d, 0755); err != nil {
			return fmt.Errorf("failed to create output directory %s: %w", d, err)
		}
	}

	cmd := exec.Command("trimmomatic", "PE", "-threads", fmt.Sprintf("%d", s.Threads), "-phred33",
		s.InputFq1, s.InputFq2,
		s.PairedOutput1, s.UnpairedOutput1,
		s.PairedOutput2, s.UnpairedOutput2,
		fmt.Sprintf("ILLUMINACLIP:%s:2:30:10", s.AdapterFastaPath),
		"LEADING:20", "TRAILING:20", "SLIDINGWINDOW:4:25", "MINLEN:30")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("trimmomatic command failed: %w", err)
	}

	if !fileExists(s.PairedOutput1) || !fileExists(s.PairedOutput2) {
		return fmt.Errorf("trimmomatic failed, expected files not found: %s, %s", s.PairedOutput1, s.PairedOutput2)
	}

	// Validate that resulting gz files are not truncated/corrupt
	if !gzipIntegrityOK(s.PairedOutput1) || !gzipIntegrityOK(s.PairedOutput2) {
		return fmt.Errorf("trimmomatic produced invalid gzip outputs (possible truncation)")
	}

	fmt.Println("Trimmomatic trimming completed.")
	return nil
}
