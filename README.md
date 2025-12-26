# bcopy

Bulk copy codebase files to clipboard with smart filtering in markdown format for feeding it to LLMs.

Perfect for sharing code context with LLMs, code reviews, or documentation.

## Features

- Works in git repos (including subfolders) and regular directories
- Respects .gitignore patterns in git repos (optional)
- Smart filtering for common artifacts and dependencies
- Binary file detection (skips files with null bytes)
- Symlink loop prevention for safe traversal
- Per-file and total size limits with safety guards
- Beautiful colored progress indicator
- Multiple output modes: clipboard, file, or stdout
- Clean markdown format with syntax highlighting for 50+ languages
- Configurable via file or flags
- Supports 50+ languages including Go, Python, JavaScript, TypeScript, Rust, and more

## Installation

```bash
# Homebrew
brew tap nodelike/tap
brew install bcopy

# Go
go install github.com/nodelike/bcopy/cmd/bcopy@latest
```

## Usage

```bash
# Basic usage
bcopy                           # Copy current dir to clipboard
bcopy ./src                     # Copy specific folder
bcopy --dry-run                 # Print to stdout
bcopy -o output.md              # Write to file

# Filtering
bcopy --exclude-tests           # Skip test files
bcopy --no-gitignore            # Ignore .gitignore
bcopy --max-depth 3             # Max 3 levels deep (default: unlimited)
bcopy --ext .go --ext .py       # Only Go and Python files

# Size limits
bcopy --threshold 5             # Warn at 5MB (default: 1MB)
bcopy --hard-max 100            # Abort at 100MB (default: 50MB)
bcopy --max-file-size 20        # Skip files >20MB (default: 10MB)
```

**Output:** Clean markdown with syntax highlighting for 50+ languages

### Config File

Create `.bcopy.yaml` in your project root (or any parent directory) for per-repo configuration:

```yaml
exclude-tests: true
no-gitignore: false
max-depth: 0
threshold: 2.0
hard-max: 100.0
max-file-size: 20.0

exclude:
  - "vendor/"
  - "\\.pb\\.go$"

ext:
  - ".go"
  - ".py"
  - ".js"
```

## Smart Filtering

**Auto-excludes:** `node_modules`, `.git`, `dist`, `build`, `vendor`, lock files, binaries, images, generated files

**Includes:** 50+ languages (Go, Python, JS/TS, Rust, Java, C/C++, Ruby, PHP, Terraform, etc.) + config files (YAML, JSON, TOML, HCL, etc.)

**Safety:** Binary detection, size limits, symlink loop prevention

## Common Use Cases

```bash
bcopy                        # Feed entire project to LLM
bcopy ./src --exclude-tests  # Just source code
bcopy --dry-run | head -n 50 # Preview output
bcopy -o review.md           # Save for code review
```

## Requirements

Clipboard support (macOS, Linux, Windows) • Works best in git repos but runs anywhere

## License

[MIT License](LICENSE)

made with ❤️ by [@nodelike](https://nodelike.com/)

