package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type PilonStep struct {
	ContigsIn      string
	TrimmedPaired1 string
	TrimmedPaired2 string
	PilonDir       string
	Threads        int
	Memory         int
	PilonJarPath   string
}

func (s *PilonStep) Name() string {
	return "Pilon Polishing"
}

func (s *PilonStep) Run() error {
	pilonContigsFile := filepath.Join(s.PilonDir, "pilon_r1.fasta")
	if fileExists(pilonContigsFile) {
		fmt.Println("Pilon corrected contigs already exist, skipping polishing.")
		return nil
	}

	fmt.Println("Running Pilon for assembly polishing...")

	// Ensure pilon output directory exists
	if err := os.MkdirAll(s.PilonDir, 0755); err != nil {
		return fmt.Errorf("failed to create Pilon output directory: %w", err)
	}

	bamFile := filepath.Join(s.PilonDir, "mapped_reads.sorted.bam")

	cmdIndex := exec.Command("bwa", "index", s.ContigsIn)
	cmdIndex.Stdout = os.Stdout
	cmdIndex.Stderr = os.Stderr
	if err := cmdIndex.Run(); err != nil {
		return fmt.Errorf("bwa index failed: %w", err)
	}

	bwaCmd := fmt.Sprintf("bwa mem -t %d %s %s %s | samtools sort -@ %d -o %s -", s.Threads, s.ContigsIn, s.TrimmedPaired1, s.TrimmedPaired2, s.Threads, bamFile)
	cmdMem := exec.Command("bash", "-c", bwaCmd)
	cmdMem.Stdout = os.Stdout
	cmdMem.Stderr = os.Stderr
	if err := cmdMem.Run(); err != nil {
		return fmt.Errorf("bwa mem and samtools sort failed: %w", err)
	}

	cmdSamIndex := exec.Command("samtools", "index", bamFile)
	cmdSamIndex.Stdout = os.Stdout
	cmdSamIndex.Stderr = os.Stderr
	if err := cmdSamIndex.Run(); err != nil {
		return fmt.Errorf("samtools index failed: %w", err)
	}

	cmdPilon := exec.Command("java", fmt.Sprintf("-Xmx%dG", s.Memory), "-jar", s.PilonJarPath,
		"--genome", s.ContigsIn, "--frags", bamFile, "--output", "pilon_r1", "--outdir", s.PilonDir,
		"--changes", "--fix", "snps,indels", "--threads", fmt.Sprintf("%d", s.Threads))
	cmdPilon.Stdout = os.Stdout
	cmdPilon.Stderr = os.Stderr
	if err := cmdPilon.Run(); err != nil {
		return fmt.Errorf("pilon command failed: %w", err)
	}

	if !fileExists(pilonContigsFile) {
		return fmt.Errorf("pilon failed, expected file not found: %s", pilonContigsFile)
	}

	fmt.Println("Pilon polishing completed.")
	return nil
}
