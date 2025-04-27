# Gemini CLI Tool

A command-line interface (CLI) tool built with Go to interact with the Google Gemini API. Send prompts, upload files, and engage in chat sessions directly from your terminal.

**Version:** 1.0.0 (Currently in Beta Testing)

## Features

* Send text prompts to the Gemini API.
* Upload files (text, code, images, etc.) along with your prompts.
* Engage in interactive chat sessions.
* Automatically detect and save code blocks from responses into files.
* Configure your API key and preferred model.
* Cross-platform (builds for Windows, macOS, Linux).

## Installation

1.  **Download the Latest Release:**
    * Go to the [Releases](https://github.com/YOUR_USERNAME/YOUR_REPOSITORY/releases) page of this repository. ( **<- Replace with your actual GitHub repo link!** )
    * Download the appropriate binary for your operating system (e.g., `gemini-cli-windows-amd64.exe`, `gemini-cli-linux-amd64`, `gemini-cli-macos-arm64`).

2.  **Place the Executable:**
    * Rename the downloaded file to something simple like `gemini-cli.exe` (Windows) or `gemini-cli` (macOS/Linux).
    * Move this executable file to a directory that is included in your system's PATH environment variable. Common locations include:
        * **Windows:** Create a folder like `C:\Program Files\gemini-cli\` or `C:\Users\YourUsername\bin\` and place the `.exe` file there.
        * **macOS/Linux:** `/usr/local/bin/` (may require `sudo`) or create a `~/bin` or `~/.local/bin` directory in your home folder and place the binary there.

3.  **Add Directory to PATH (if needed):**
    * If the directory you chose in step 2 is not already in your PATH, you need to add it. Search online for specific instructions for "add directory to PATH" for your operating system (Windows, macOS, or your Linux distribution).
    * **Example (Linux/macOS - Bash/Zsh):** Add `export PATH="/path/to/your/cli/directory:$PATH"` to your `~/.bashrc`, `~/.zshrc`, or `~/.profile` file. Replace `/path/to/your/cli/directory` with the actual path (e.g., `~/bin`). Remember to run `source ~/.bashrc` (or the relevant file) or restart your terminal.
    * **Example (Windows):** Search for "Environment Variables", edit "System variables" or "User variables", find `Path`, click "Edit", then "New", and add the full path to the directory (e.g., `C:\Program Files\gemini-cli\`). Restart your terminal.

4.  **Verify Installation:**
    * Open a *new* terminal or command prompt window.
    * Type `gemini-cli --version` and press Enter.
    * You should see `Gemini CLI Tool Version: 1.0.0` (or the current version).

## Configuration

Before using the tool for the first time, you need to initialize it to save your Gemini API key.

```bash
gemini-cli init
