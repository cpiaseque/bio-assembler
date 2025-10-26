package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type SpadesStep struct {
	InputFq1 string
	InputFq2 string
	Output   string
	Threads  int
	Memory   int
}

func (s *SpadesStep) Name() string {
	return "SPAdes Assembly"
}

func (s *SpadesStep) Run() error {
	contigsFile := filepath.Join(s.Output, "contigs.fasta")
	if fileExists(contigsFile) {
		fmt.Println("SPAdes contigs already exist, skipping assembly.")
		return nil
	}

	fmt.Println("Running SPAdes for de novo assembly...")

	// Ensure output directory exists
	if err := os.MkdirAll(s.Output, 0755); err != nil {
		return fmt.Errorf("failed to create SPAdes output directory: %w", err)
	}

	cmd := exec.Command("spades.py", "--only-assembler", "--careful",
		"-t", fmt.Sprintf("%d", s.Threads),
		"-m", fmt.Sprintf("%d", s.Memory),
		"--pe1-1", s.InputFq1,
		"--pe1-2", s.InputFq2,
		"-o", s.Output)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("spades command failed: %w", err)
	}

	if !fileExists(contigsFile) {
		return fmt.Errorf("spades failed, expected file not found: %s", contigsFile)
	}

	fmt.Println("SPAdes assembly completed.")
	return nil
}
