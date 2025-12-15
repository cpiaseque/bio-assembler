package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bio-assembler/pkg/pipeline"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

var (
	srrID            string
	threads          int
	memory           int
	pilonJarPath     string
	adapterFastaPath string
	noParallel       bool
	filterMode       string
	filterCustomArgs string
)

func init() {
	runCmd.Flags().StringVarP(&srrID, "srr", "s", "", "SRR ID of the sample to process (required)")
	runCmd.Flags().IntVarP(&threads, "threads", "t", 4, "Number of threads to use")
	runCmd.Flags().IntVarP(&memory, "memory", "m", 16, "Memory in GB to use")
	runCmd.Flags().StringVar(&pilonJarPath, "pilon-jar", "", "Path to the pilon.jar file (required)")
	runCmd.Flags().StringVar(&adapterFastaPath, "adapter-fasta", "", "Path to the adapter FASTA file for Trimmomatic (required)")
	runCmd.Flags().BoolVar(&noParallel, "no-parallel", false, "Disable parallel execution where possible")
	runCmd.Flags().StringVar(&filterMode, "filter-mode", "standard", "Read filtering mode for Trimmomatic: standard, strict, lenient, or custom")
	runCmd.Flags().StringVar(&filterCustomArgs, "filter-custom-args", "", "Custom Trimmomatic filtering arguments (used only when --filter-mode=custom)")

	runCmd.MarkFlagRequired("srr")
	runCmd.MarkFlagRequired("pilon-jar")
	runCmd.MarkFlagRequired("adapter-fasta")

	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the full genome assembly pipeline",
	Run: func(cmd *cobra.Command, args []string) {
		if srrID == "" {
			log.Fatal("SRR ID must be provided.")
		}

		baseDir, err := os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current working directory: %v", err)
		}
		sampleDir := filepath.Join(baseDir, "data", srrID)
		rawDir := filepath.Join(sampleDir, "raw_data")
		fastqcRawDir := filepath.Join(sampleDir, "01_fastqc_raw")
		trimmedDir := filepath.Join(sampleDir, "02_trimmed_reads")
		fastqcTrimmedDir := filepath.Join(sampleDir, "03_fastqc_trimmed")
		spadesDir := filepath.Join(sampleDir, "04_spades_assembly")
		pilonDir := filepath.Join(sampleDir, "05_pilon_correction", "round1")
		qualimapDir := filepath.Join(sampleDir, "08_qualimap_report")

		rawFq1 := filepath.Join(rawDir, srrID+"_1.fastq.gz")
		rawFq2 := filepath.Join(rawDir, srrID+"_2.fastq.gz")
		trimmedPaired1 := filepath.Join(trimmedDir, "trimmed_paired_1.fastq.gz")
		trimmedPaired2 := filepath.Join(trimmedDir, "trimmed_paired_2.fastq.gz")
		trimmedUnpaired1 := filepath.Join(trimmedDir, "trimmed_unpaired_1.fastq.gz")
		trimmedUnpaired2 := filepath.Join(trimmedDir, "trimmed_unpaired_2.fastq.gz")
		spadesContigs := filepath.Join(spadesDir, "contigs.fasta")
		pilonContigs := filepath.Join(pilonDir, "pilon_r1.fasta")
		bamFile := filepath.Join(pilonDir, "mapped_reads.sorted.bam")

		// Construct step instances
		download := &pipeline.DownloadStep{
			SrrID:   srrID,
			Output:  rawDir,
			Threads: threads,
		}
		fastqcRaw := &pipeline.FastQCStep{
			InputFq1: rawFq1,
			InputFq2: rawFq2,
			Output:   fastqcRawDir,
			Threads:  threads,
		}
		trim := &pipeline.TrimmomaticStep{
			InputFq1:         rawFq1,
			InputFq2:         rawFq2,
			PairedOutput1:    trimmedPaired1,
			PairedOutput2:    trimmedPaired2,
			UnpairedOutput1:  trimmedUnpaired1,
			UnpairedOutput2:  trimmedUnpaired2,
			Threads:          threads,
			AdapterFastaPath: adapterFastaPath,
		Mode:             filterMode,
		CustomArgs:       filterCustomArgs,
		}
		fastqcTrim := &pipeline.TrimmedFastQCStep{
			InputFq1: trimmedPaired1,
			InputFq2: trimmedPaired2,
			Output:   fastqcTrimmedDir,
			Threads:  threads,
		}
		spades := &pipeline.SpadesStep{
			InputFq1: trimmedPaired1,
			InputFq2: trimmedPaired2,
			Output:   spadesDir,
			Threads:  threads,
			Memory:   memory,
		}
		pilon := &pipeline.PilonStep{
			ContigsIn:      spadesContigs,
			TrimmedPaired1: trimmedPaired1,
			TrimmedPaired2: trimmedPaired2,
			PilonDir:       pilonDir,
			Threads:        threads,
			Memory:         memory,
			PilonJarPath:   pilonJarPath,
		}
		qualimap := &pipeline.QualimapStep{
			BamFile:   bamFile,
			OutputDir: qualimapDir,
			Memory:    memory,
		}

		// Phase 1: Download (sequential)
		fmt.Println("=== RUNNING STEP: ", download.Name(), "===")
		if err := download.Run(); err != nil {
			log.Fatalf("Pipeline failed: %v", err)
		}
		fmt.Println("=== COMPLETED STEP: ", download.Name(), "===\n")

		// Phase 2: FastQC Raw and Trimmomatic in parallel (or sequential if disabled)
		if noParallel {
			fmt.Println("=== RUNNING STEP: ", fastqcRaw.Name(), "===")
			if err := fastqcRaw.Run(); err != nil {
				log.Fatalf("Pipeline failed: %v", err)
			}
			fmt.Println("=== COMPLETED STEP: ", fastqcRaw.Name(), "===\n")

			fmt.Println("=== RUNNING STEP: ", trim.Name(), "===")
			if err := trim.Run(); err != nil {
				log.Fatalf("Pipeline failed: %v", err)
			}
			fmt.Println("=== COMPLETED STEP: ", trim.Name(), "===\n")
		} else {
			fmt.Println("=== RUNNING IN PARALLEL: ", fastqcRaw.Name(), " & ", trim.Name(), "===")
			g, _ := errgroup.WithContext(context.Background())
			g.Go(func() error { return fastqcRaw.Run() })
			g.Go(func() error { return trim.Run() })
			if err := g.Wait(); err != nil {
				log.Fatalf("Pipeline failed: %v", err)
			}
			fmt.Println("=== COMPLETED PARALLEL GROUP ===\n")
		}

		// Phase 3: FastQC Trimmed then SPAdes (sequential to avoid concurrent reads of the same files)
		fmt.Println("=== RUNNING STEP: ", fastqcTrim.Name(), "===")
		if err := fastqcTrim.Run(); err != nil {
			log.Fatalf("Pipeline failed: %v", err)
		}
		fmt.Println("=== COMPLETED STEP: ", fastqcTrim.Name(), "===")
		fmt.Println()

		fmt.Println("=== RUNNING STEP: ", spades.Name(), "===")
		if err := spades.Run(); err != nil {
			log.Fatalf("Pipeline failed: %v", err)
		}
		fmt.Println("=== COMPLETED STEP: ", spades.Name(), "===")
		fmt.Println()

		// Phase 4: Pilon (sequential)
		fmt.Println("=== RUNNING STEP: ", pilon.Name(), "===")
		if err := pilon.Run(); err != nil {
			log.Fatalf("Pipeline failed: %v", err)
		}
		fmt.Println("=== COMPLETED STEP: ", pilon.Name(), "===")
		fmt.Println()

		// Phase 5: Qualimap (sequential)
		fmt.Println("=== RUNNING STEP: ", qualimap.Name(), "===")
		if err := qualimap.Run(); err != nil {
			log.Fatalf("Pipeline failed: %v", err)
		}
		fmt.Println("=== COMPLETED STEP: ", qualimap.Name(), "===")
		fmt.Println()

		fmt.Println("=======================================================================")
		fmt.Printf("Genome assembly %s complete!\n", srrID)
		fmt.Printf("Final report path: %s\n", pilonContigs)
		fmt.Printf("Qualimap report: %s/qualimap_report.html\n", qualimapDir)
		fmt.Println("=======================================================================")
	},
}
