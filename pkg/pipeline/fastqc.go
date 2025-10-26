package pipeline

import (
	"fmt"
	"os"
	"os/exec"
)

type FastQCStep struct {
	InputFq1 string
	InputFq2 string
	Output   string
	Threads  int
}

func (s *FastQCStep) Name() string {
	return "FastQC Analysis"
}

func (s *FastQCStep) Run() error {
	fmt.Println("Running FastQC for initial quality control...")
	if err := os.MkdirAll(s.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command("fastqc", s.InputFq1, s.InputFq2, "-o", s.Output, "-t", fmt.Sprintf("%d", s.Threads))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fastqc command failed: %w", err)
	}

	fmt.Println("FastQC analysis completed.")
	return nil
}

type TrimmedFastQCStep struct {
	InputFq1 string
	InputFq2 string
	Output   string
	Threads  int
}

func (s *TrimmedFastQCStep) Name() string {
	return "FastQC Analysis on Trimmed Reads"
}

func (s *TrimmedFastQCStep) Run() error {
	fmt.Println("Running FastQC for trimmed reads...")
	if err := os.MkdirAll(s.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	cmd := exec.Command("fastqc", s.InputFq1, s.InputFq2, "-o", s.Output, "-t", fmt.Sprintf("%d", s.Threads))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fastqc command failed: %w", err)
	}

	fmt.Println("FastQC analysis on trimmed reads completed.")
	return nil
}
