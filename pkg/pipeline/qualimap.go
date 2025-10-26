package pipeline

import (
	"fmt"
	"os"
	"os/exec"
)

type QualimapStep struct {
	BamFile   string
	OutputDir string
	Memory    int
}

func (s *QualimapStep) Name() string {
	return "Qualimap Quality Assessment"
}

func (s *QualimapStep) Run() error {
	fmt.Println("Running Qualimap for quality assessment...")

	// Ensure output directory exists
	if err := os.MkdirAll(s.OutputDir, 0755); err != nil {
		return fmt.Errorf("failed to create Qualimap output directory: %w", err)
	}

	cmd := exec.Command("qualimap", "bamqc", "-bam", s.BamFile, "-outdir", s.OutputDir, fmt.Sprintf("--java-mem-size=%dG", s.Memory))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("qualimap command failed: %w", err)
	}

	fmt.Println("Qualimap quality assessment completed.")
	return nil
}
