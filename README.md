# Bio Assembler CLI

A command-line tool to automate the genome assembly pipeline.

## Installation

The easiest way to install all the required bioinformatics tools is by using `conda`.

1. **Install Miniconda:** If you don't have `conda` installed, follow the instructions on the [Miniconda website](https://docs.conda.io/en/latest/miniconda.html).

2. **Set up Bioconda channels:** Before installing the tools, you need to configure the correct channels. This only needs to be done once.

    ```bash
    conda config --add channels defaults
    conda config --add channels bioconda
    conda config --add channels conda-forge
    ```

3. **Create a Conda environment:** This command will create a new environment named `bio-assembler` and install all the necessary tools.

    ```bash
    conda create -n bio-assembler -c bioconda -c conda-forge fastqc sra-tools trimmomatic spades bwa samtools qualimap pilon
    ```

4. **Activate the environment:** Before running the `bio-assembler` CLI, you must activate the conda environment:

    ```bash
    conda activate bio-assembler
    ```

## Build

To build the application, run:

```bash
go build -o bio-assembler ./cmd/bio-assembler
```

## Run

To execute the full pipeline for a given sample:

```bash
./bio-assembler run \
  -s <SRR_ID> \
  --pilon-jar /path/to/pilon.jar \
  --adapter-fasta /path/to/adapters.fa
```

Example:

```bash
./bio-assembler run \
  -s SRR13511998 \
  --pilon-jar /home/user/tools/pilon-1.24.jar \
  --adapter-fasta /home/user/tools/Trimmomatic-0.39/adapters/TruSeq3-PE.fa
```

## Dependencies

### Pilon

Pilon is a tool to polish the genome assembly.

* **Download:** You can download the `pilon.jar` file from the [official Pilon GitHub repository](https://github.com/broadinstitute/pilon/releases).
* **Usage:** Provide the full path to the downloaded `pilon-X.Y.Z.jar` file using the `--pilon-jar` flag.

### Trimmomatic Adapters

The adapter file is required by Trimmomatic to remove sequencing adapters from the raw reads.

* **Location:** The adapter FASTA files are included with your Trimmomatic installation, usually in an `adapters/` subdirectory.
* **How to choose:**
    1. **FastQC Analysis (Recommended):** Run a FastQC analysis on your raw data and check the "Overrepresented sequences" report. This is the most reliable method as it shows what is actually in your data.
    2. **NCBI SRA Database:** Search for your SRR ID on the [NCBI SRA website](https://www.ncbi.nlm.nih.gov/sra). In the experiment details, look for the "Library Preparation Kit". The kit's name (e.g., "Illumina TruSeq") tells you which adapters were used.
    3. **Common Default:** For most modern Illumina paired-end data, `TruSeq3-PE.fa` is the correct file.
* **Usage:** Provide the full path to the appropriate adapter file (e.g., `TruSeq3-PE.fa`) using the `--adapter-fasta` flag.

## Troubleshooting

If you encounter issues like `conda: command not found`, please refer to the [TROUBLESHOOTING.md](TROUBLESHOOTING.md) file for solutions.
