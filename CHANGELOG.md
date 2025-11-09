# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-01-09

### Initial Release

**Core Features:**
- Bulk copy codebase files to clipboard in markdown format
- Designed specifically for feeding code context to LLMs
- Works in git repositories (including subfolders)
- Git subdirectory support with automatic parent directory traversal
- Respects .gitignore patterns (configurable)

**Smart Filtering:**
- 50+ language support (Go, Python, JS/TS, Rust, Java, C/C++, Ruby, PHP, Perl, Lua, Elixir, Erlang, Clojure, Dart, R, Terraform, and more)
- Auto-excludes common artifacts (node_modules, dist, build, vendor, .git, etc.)
- Binary file detection (null byte check in first 8KB)
- Symlink loop prevention for safe traversal
- Configurable file extension filtering

**Output Modes:**
- Copy to clipboard (default)
- `--dry-run`: Print to stdout for piping/preview
- `-o, --output`: Write to file

**Safety Features:**
- Path validation (blocks root, home, system directories)
- Hard maximum size limit (default 50MB, configurable)
- Per-file size limit (default 10MB, configurable)
- Warning threshold with user confirmation (default 1MB, configurable)
- Binary file detection and skipping
- Symlink loop prevention
- Graceful interrupt handling (Ctrl+C)

**Configuration:**
- Command-line flags for all options
- `.bcopy.yaml` config file support
- Viper-based configuration management

**UI/UX:**
- Beautiful colored progress indicator with dots
- Clean, informative output messages
- File count and size reporting
- Success/error status with emojis

**Flags:**
- `--no-gitignore`: Ignore .gitignore patterns
- `--exclude-tests`: Exclude test files
- `--exclude`: Custom exclusion patterns (repeatable)
- `--ext`: Override allowed extensions (repeatable)
- `--max-depth`: Limit directory traversal depth
- `--threshold`: Size warning threshold in MB
- `--hard-max`: Hard maximum size in MB (aborts if exceeded)
- `--max-file-size`: Maximum individual file size in MB
- `--dry-run`: Print to stdout instead of clipboard
- `-o, --output`: Write to file instead of clipboard
- `--config`: Custom config file path

**Distribution:**
- Homebrew tap support (`brew install nodelike/tap/bcopy`)
- Go install support (`go install github.com/nodelike/bcopy/cmd/bcopy@latest`)
- Automated releases via GitHub Actions and GoReleaser
- Pre-built binaries for Linux and macOS (amd64, arm64)

### Technical Details
- Built with Go 1.24.4
- Uses go-git for repository detection
- Cross-platform clipboard support via atotto/clipboard
- Concurrent file processing with errgroup
- Regex-based filtering with gobwas/glob for gitignore patterns
- Clean codebase with no external UI dependencies

