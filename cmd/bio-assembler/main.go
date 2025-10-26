package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "bio-assembler",
	Short: "A CLI tool for bioinformatics genome assembly.",
	Long: `bio-assembler is a command-line tool to automate the process of
genome assembly from raw sequencing reads. It includes steps for data download,
quality control, trimming, assembly, and polishing.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Bio Assembler v0.1.0")
		fmt.Println("Use 'bio-assembler help' for a list of commands.")
	},
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
