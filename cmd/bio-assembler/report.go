package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reportCmd)
}

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a report for a completed assembly",
	Run: func(cmd *cobra.Command, args []string) {
		if srrID == "" {
			fmt.Println("SRR ID must be provided with -s flag.")
			os.Exit(1)
		}

		baseDir, _ := os.Getwd()
		sampleDir := filepath.Join(baseDir, "data", srrID)

		fmt.Printf("Generating report for sample %s\n\n", srrID)

		fmt.Println("--- Step 1: Initial Quality Control (FastQC) ---")
		fmt.Printf("FastQC reports are in: %s\n", filepath.Join(sampleDir, "01_fastqc_raw"))
		fmt.Println("ACTION: Open the HTML reports, take screenshots of 'Per base sequence quality' graphs.")
		fmt.Println("SUGGESTED TEXT: 'Исходные данные показали падение качества к концам прочтений, что характерно для технологии Illumina. Также возможно наличие адаптерных последовательностей.'")
		prompt()

		fmt.Println("--- Step 2: Post-Trimming Quality Control (FastQC) ---")
		fmt.Printf("Trimmed FastQC reports are in: %s\n", filepath.Join(sampleDir, "03_fastqc_trimmed"))
		fmt.Println("ACTION: Open the HTML reports for trimmed data, take screenshots of 'Per base sequence quality' graphs.")
		fmt.Println("SUGGESTED TEXT: 'После очистки с помощью Trimmomatic... качество прочтений значительно улучшилось.'")
		prompt()

		fmt.Println("--- Step 3: Assembly Statistics (SPAdes) ---")
		contigsFile := filepath.Join(sampleDir, "04_spades_assembly", "contigs.fasta")
		count, _ := countFastaContigs(contigsFile)
		fmt.Printf("SPAdes assembled %d contigs.\n", count)
		fmt.Printf("SUGGESTED TEXT: 'Сборка de novo проводилась с помощью ассемблера SPAdes. В результате был получен черновой геном, состоящий из %d контигов.'\n", count)
		prompt()

		fmt.Println("--- Step 4: Polishing Statistics (Pilon) ---")
		pilonContigsFile := filepath.Join(sampleDir, "05_pilon_correction", "round1", "pilon_r1.fasta")
		pilonChangesFile := filepath.Join(sampleDir, "05_pilon_correction", "round1", "pilon_r1.changes")
		pilonCount, _ := countFastaContigs(pilonContigsFile)
		pilonChanges, _ := countLines(pilonChangesFile)
		fmt.Printf("Pilon polishing resulted in %d contigs.\n", pilonCount)
		fmt.Printf("Pilon made %d changes.\n", pilonChanges)
		fmt.Printf("SUGGESTED TEXT: 'Черновая сборка была отфильтрована... Затем с помощью Pilon было исправлено %d ошибок...'\n", pilonChanges)
		prompt()

		fmt.Println("--- Step 5: Final Quality Assessment (Qualimap) ---")
		qualimapReport := filepath.Join(sampleDir, "08_qualimap_report", "qualimap_report.html")
		fmt.Println("ACTION: Open the Qualimap report:", qualimapReport)
		fmt.Println("ACTION: Get N50 value from the prinseq output during the run.")
		fmt.Println("ACTION: Take screenshots of 'Summary' (for mean coverage) and 'Coverage across reference' graphs.")
		fmt.Println("SUGGESTED TEXT: 'Финальная сборка генома... имеет общую длину Z Mb, состоит из X контигов с N50 равным W bp... Среднее покрытие составило V-x...'")
		prompt()

		fmt.Println("Report generation guide finished.")
	},
}

func prompt() {
	fmt.Print("Press [Enter] to continue...")
	fmt.Scanln()
	fmt.Println()
}

func countFastaContigs(filePath string) (int, error) {
	cmd := exec.Command("grep", "-c", ">", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var count int
	fmt.Sscanf(string(output), "%d", &count)
	return count, nil
}

func countLines(filePath string) (int, error) {
	cmd := exec.Command("wc", "-l", filePath)
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	var count int
	fmt.Sscanf(string(output), "%d", &count)
	return count, nil
}
