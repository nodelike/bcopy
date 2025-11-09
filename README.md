# bcopy

Bulk copy codebase files to clipboard with smart filtering in markdown format for feeding it to LLMs.

Perfect for sharing code context with LLMs, code reviews, or documentation.

## Features

- Works in git repositories (including subfolders)
- Respects .gitignore patterns (optional)
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
brew install nodelike/tap/bcopy

# Go
go install github.com/nodelike/bcopy/cmd/bcopy@latest
```

## Usage

```bash
bcopy                     # Copy current directory to clipboard
bcopy ./src               # Copy specific subfolder
bcopy --exclude-tests     # Exclude test files
bcopy --dry-run           # Print to stdout instead of clipboard
bcopy -o output.md        # Write to file
bcopy --threshold 5       # Set 5MB warning threshold
bcopy --hard-max 100      # Set 100MB hard limit
bcopy --max-file-size 20  # Skip files larger than 20MB
bcopy --no-gitignore      # Ignore .gitignore patterns
```

### Output Format

Clean markdown with syntax highlighting for 50+ languages. Each file is formatted as:
```
File: ./path/to/file.go

​```go
// file contents
​```
```

### Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--no-gitignore` | Ignore .gitignore patterns | false |
| `--exclude-tests` | Exclude test files | false |
| `--exclude <pattern>` | Custom exclusion regex (repeatable) | - |
| `--ext <ext>` | Override allowed extensions (repeatable) | 50+ defaults |
| `--max-depth <n>` | Limit directory depth (0 = unlimited) | 0 |
| `--threshold <n>` | Size warning threshold in MB | 1.0 |
| `--hard-max <n>` | Hard maximum total size in MB (aborts) | 50.0 |
| `--max-file-size <n>` | Maximum individual file size in MB | 10.0 |
| `--dry-run` | Print to stdout instead of clipboard | false |
| `-o, --output <file>` | Write to file instead of clipboard | - |
| `--config <file>` | Config file path | .bcopy.yaml |

### Config File

Create `.bcopy.yaml` in your project root:

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

**Includes:** 50+ languages (Go, Python, JS/TS, Rust, Java, C/C++, Ruby, PHP, etc.) + config files (YAML, JSON, TOML, etc.)

**Safety:** Binary detection, size limits, symlink loop prevention

## Common Use Cases

```bash
bcopy                        # Feed entire project to LLM
bcopy ./src --exclude-tests  # Just source code
bcopy --dry-run | head -n 50 # Preview output
bcopy -o review.md           # Save for code review
```

## Requirements

Git repository • Clipboard support (macOS, Linux, Windows)

## License

[MIT License](LICENSE)

made with ❤️ by [@nodelike](https://nodelike.com/)

