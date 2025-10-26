## Troubleshooting

### `conda: command not found`

If you have installed Anaconda/Miniconda but the `conda` command is not found in your terminal, you need to initialize your shell.

1. Find your conda installation path. It's usually `~/anaconda3` or `~/miniconda3`.
2. Run the initialization command for your shell. For `zsh`, the command is:

    ```bash
    /path/to/your/conda/bin/conda init zsh
    ```

    For example:

    ```bash
    ~/anaconda3/bin/conda init zsh
    ```

3. Restart your terminal.

After these steps, the `conda` command should be available.
