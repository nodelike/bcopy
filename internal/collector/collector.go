package collector

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/nodelike/bcopy/internal/analyzer"
	"golang.org/x/sync/errgroup"
)

type FileData struct {
	RelPath  string
	Content  string
	Size     int64
	Language string
}

type CollectionResult struct {
	Files     []FileData
	TotalSize int64
	FileCount int
}

func Collect(ctx context.Context, rootPath string, filter *analyzer.Filter, maxDepth int, maxFileSizeMB float64) (*CollectionResult, error) {
	result := &CollectionResult{
		Files: make([]FileData, 0),
	}

	type fileJob struct {
		fullPath string
		relPath  string
	}

	fileJobs := make([]fileJob, 0)
	visitedDirs := make(map[string]bool) // Track visited directories to avoid symlink loops

	err := filepath.WalkDir(rootPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		// Handle symlinks to avoid infinite loops
		if d.Type()&os.ModeSymlink != 0 {
			realPath, err := filepath.EvalSymlinks(path)
			if err != nil {
				return nil // Skip broken symlinks
			}

			// Check if we've already visited this real path
			if visitedDirs[realPath] {
				return filepath.SkipDir // Skip to avoid loop
			}

			// Mark as visited if it's a directory
			info, err := os.Stat(realPath)
			if err == nil && info.IsDir() {
				visitedDirs[realPath] = true
			}
		}

		relPath, err := filepath.Rel(rootPath, path)
		if err != nil {
			return nil
		}

		if d.IsDir() {
			if relPath != "." {
				depth := strings.Count(relPath, string(os.PathSeparator)) + 1
				if maxDepth > 0 && depth > maxDepth {
					return filepath.SkipDir
				}

				if !filter.ShouldInclude(relPath + "/dummy.go") {
					return filepath.SkipDir
				}
			}
			return nil
		}

		if !filter.ShouldInclude(relPath) {
			return nil
		}

		fileJobs = append(fileJobs, fileJob{fullPath: path, relPath: relPath})
		return nil
	})

	if err != nil {
		return nil, err
	}

	eg, egCtx := errgroup.WithContext(ctx)
	eg.SetLimit(16)

	type fileResult struct {
		data FileData
		err  error
	}

	resultsChan := make(chan fileResult, len(fileJobs))

	// Simple progress without terminal manipulation
	fmt.Fprintf(os.Stderr, "\033[36m📦 Collecting files...\033[0m ")

	progressDots := 0
	progressTicker := make(chan struct{}, 10)

	go func() {
		for range progressTicker {
			if progressDots < 3 {
				fmt.Fprint(os.Stderr, ".")
				progressDots++
			}
		}
	}()

	for _, job := range fileJobs {
		job := job
		eg.Go(func() error {
			select {
			case <-egCtx.Done():
				return egCtx.Err()
			default:
			}

			// Check if file is binary by reading first chunk
			if isBinary, err := isBinaryFile(job.fullPath); err != nil || isBinary {
				select {
				case progressTicker <- struct{}{}:
				default:
				}
				return nil // Skip binary files
			}

			info, err := os.Stat(job.fullPath)
			if err != nil {
				resultsChan <- fileResult{err: err}
				select {
				case progressTicker <- struct{}{}:
				default:
				}
				return nil
			}

			// Check file size limit
			fileSizeMB := float64(info.Size()) / (1024 * 1024)
			if maxFileSizeMB > 0 && fileSizeMB > maxFileSizeMB {
				select {
				case progressTicker <- struct{}{}:
				default:
				}
				return nil // Skip files that are too large
			}

			content, err := os.ReadFile(job.fullPath)
			if err != nil {
				resultsChan <- fileResult{err: err}
				select {
				case progressTicker <- struct{}{}:
				default:
				}
				return nil
			}

			fileData := FileData{
				RelPath:  job.relPath,
				Content:  string(content),
				Size:     info.Size(),
				Language: getLanguage(job.relPath),
			}

			resultsChan <- fileResult{data: fileData}
			select {
			case progressTicker <- struct{}{}:
			default:
			}
			return nil
		})
	}

	go func() {
		eg.Wait()
		close(resultsChan)
	}()

	for res := range resultsChan {
		if res.err != nil {
			continue
		}
		result.Files = append(result.Files, res.data)
		result.TotalSize += res.data.Size
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	close(progressTicker)
	fmt.Fprintf(os.Stderr, " \033[32m✓\033[0m (%d files)\n", len(result.Files))

	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].RelPath < result.Files[j].RelPath
	})

	result.FileCount = len(result.Files)

	return result, nil
}

func FormatAsMarkdown(result *CollectionResult) string {
	var sb strings.Builder

	for i, file := range result.Files {
		sb.WriteString(fmt.Sprintf("File: ./%s\n\n", file.RelPath))
		sb.WriteString(fmt.Sprintf("```%s\n", file.Language))
		sb.WriteString(file.Content)
		if !strings.HasSuffix(file.Content, "\n") {
			sb.WriteString("\n")
		}
		sb.WriteString("```\n")

		if i < len(result.Files)-1 {
			sb.WriteString("\n---\n\n")
		}
	}

	return sb.String()
}

// isBinaryFile checks if a file is binary by reading the first chunk
// Returns true if the file contains null bytes (binary indicator)
func isBinaryFile(filePath string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	// Read first 8KB to check for binary content
	buf := make([]byte, 8192)
	n, err := file.Read(buf)
	if err != nil && n == 0 {
		return false, err
	}

	// Check for null bytes which indicate binary content
	return bytes.IndexByte(buf[:n], 0) != -1, nil
}

func getLanguage(filename string) string {
	base := filepath.Base(filename)
	ext := filepath.Ext(filename)

	// Check for special files without extensions
	noExtMap := map[string]string{
		"Makefile":    "makefile",
		"Dockerfile":  "dockerfile",
		"Rakefile":    "ruby",
		"Gemfile":     "ruby",
		"Procfile":    "yaml",
		"Vagrantfile": "ruby",
		"Cargo":       "toml",
	}

	if ext == "" {
		if lang, ok := noExtMap[base]; ok {
			return lang
		}
		return ""
	}

	languageMap := map[string]string{
		".go":         "go",
		".py":         "python",
		".js":         "javascript",
		".jsx":        "jsx",
		".ts":         "typescript",
		".tsx":        "tsx",
		".vue":        "vue",
		".svelte":     "svelte",
		".mjs":        "javascript",
		".cjs":        "javascript",
		".rs":         "rust",
		".java":       "java",
		".c":          "c",
		".cpp":        "cpp",
		".cc":         "cpp",
		".cxx":        "cpp",
		".h":          "c",
		".hpp":        "cpp",
		".hh":         "cpp",
		".cs":         "csharp",
		".rb":         "ruby",
		".php":        "php",
		".swift":      "swift",
		".kt":         "kotlin",
		".kts":        "kotlin",
		".scala":      "scala",
		".sh":         "bash",
		".bash":       "bash",
		".zsh":        "zsh",
		".fish":       "fish",
		".yaml":       "yaml",
		".yml":        "yaml",
		".json":       "json",
		".xml":        "xml",
		".html":       "html",
		".htm":        "html",
		".css":        "css",
		".scss":       "scss",
		".sass":       "sass",
		".less":       "less",
		".md":         "markdown",
		".markdown":   "markdown",
		".sql":        "sql",
		".toml":       "toml",
		".ini":        "ini",
		".conf":       "conf",
		".env":        "bash",
		".txt":        "text",
		".dockerfile": "dockerfile",
		".pl":         "perl",
		".pm":         "perl",
		".lua":        "lua",
		".vim":        "vim",
		".ex":         "elixir",
		".exs":        "elixir",
		".erl":        "erlang",
		".hrl":        "erlang",
		".clj":        "clojure",
		".cljs":       "clojure",
		".dart":       "dart",
		".r":          "r",
		".R":          "r",
		".m":          "objective-c",
		".mm":         "objective-c",
		".groovy":     "groovy",
		".gradle":     "gradle",
		".tf":         "terraform",
		".hcl":        "hcl",
	}

	if lang, ok := languageMap[ext]; ok {
		return lang
	}
	return ""
}
