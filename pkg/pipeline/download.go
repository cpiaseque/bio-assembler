package pipeline

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type DownloadStep struct {
	SrrID   string
	Output  string
	Threads int
}

func (s *DownloadStep) Name() string {
	return "Download Raw Data"
}

func (s *DownloadStep) Run() error {
	rawFq1 := filepath.Join(s.Output, s.SrrID+"_1.fastq.gz")
	rawFq2 := filepath.Join(s.Output, s.SrrID+"_2.fastq.gz")

	if fileExists(rawFq1) && fileExists(rawFq2) {
		fmt.Printf("Files for %s already exist, skipping download.\n", s.SrrID)
		return nil
	}

	fmt.Printf("Downloading data for SRR ID: %s\n", s.SrrID)
	if err := os.MkdirAll(s.Output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Use manual progress rendering instead of relying on fastq-dump --progress
	stop := StartSpinner(fmt.Sprintf("Downloading %s (fastq-dump)", s.SrrID))

	cmd := exec.Command("fastq-dump", "--split-files", "--gzip", "-O", s.Output, s.SrrID)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	runErr := cmd.Run()
	if stop != nil {
		if runErr != nil {
			stop("failed")
		} else {
			stop("done")
		}
	}
	if runErr != nil {
		return fmt.Errorf("fastq-dump command failed: %w", runErr)
	}

	if !fileExists(rawFq1) || !fileExists(rawFq2) {
		return fmt.Errorf("download failed, expected files not found: %s, %s", rawFq1, rawFq2)
	}

	fmt.Println("Data download completed.")
	return nil
}
