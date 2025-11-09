package analyzer

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gobwas/glob"
)

type Filter struct {
	allowedExts      map[string]bool
	excludePatterns  []*regexp.Regexp
	gitignoreGlobs   []glob.Glob
	respectGitignore bool
	excludeTests     bool
}

func NewFilter(allowedExts []string, customExcludes []string, respectGitignore bool, excludeTests bool) *Filter {
	f := &Filter{
		allowedExts:      make(map[string]bool),
		respectGitignore: respectGitignore,
		excludeTests:     excludeTests,
	}

	if len(allowedExts) == 0 {
		defaultExts := []string{
			".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".vue", ".svelte", ".mjs", ".cjs",
			".yaml", ".yml", ".json", ".toml", ".md", ".txt", ".sh", ".bash",
			".c", ".cpp", ".h", ".hpp", ".rs", ".java", ".rb", ".php", ".swift", ".kt",
		}
		for _, ext := range defaultExts {
			f.allowedExts[ext] = true
		}
	} else {
		for _, ext := range allowedExts {
			if !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			f.allowedExts[ext] = true
		}
	}

	alwaysExclude := []string{
		`(^|/)node_modules($|/)`,
		`(^|/)venv($|/)`,
		`(^|/)\.venv($|/)`,
		`(^|/)__pycache__($|/)`,
		`(^|/)\.git($|/)`,
		`(^|/)dist($|/)`,
		`(^|/)build($|/)`,
		`(^|/)\.egg-info($|/)`,
		`(^|/)\.tox($|/)`,
		`(^|/)coverage($|/)`,
		`(^|/)\.next($|/)`,
		`(^|/)vendor($|/)`,
		`(^|/)bin($|/)`,
		`(^|/)tmp($|/)`,
		`\.lock$`,
		`-lock\.json$`,
		`-lock\.yaml$`,
		`Pipfile\.lock$`,
		`\.gitignore$`,
		`\.exe$`,
		`\.so$`,
		`\.dylib$`,
		`\.dll$`,
		`_templ\.go$`,
		`\.(jpg|jpeg|png|gif|bmp|svg|ico|webp|tiff|tif|psd|raw|heic|avif)$`,
		`\.pyc$`,
		`\.pyo$`,
		`\.pyd$`,
		`\.egg$`,
		`(^|/)\.eggs($|/)`,
		`(^|/)\.pytest_cache($|/)`,
		`(^|/)\.mypy_cache($|/)`,
		`\.pb\.go$`,
		`_gen\.go$`,
		`\.min\.js$`,
		`\.bundle\.js$`,
		`\.eslintcache`,
		`(^|/)\.nyc_output($|/)`,
		`(^|/)\.yarn($|/)`,
		`(^|/)\.npm($|/)`,
		`(^|/)cypress($|/)`,
		`(^|/)jest-cache($|/)`,
	}

	testPatterns := []string{
		`_test\.go$`,
		`(^|/)tests?($|/)`,
		`\.test\.(js|ts|jsx|tsx)$`,
		`\.spec\.(js|ts|jsx|tsx)$`,
	}

	allPatterns := alwaysExclude
	if f.excludeTests {
		allPatterns = append(allPatterns, testPatterns...)
	}
	allPatterns = append(allPatterns, customExcludes...)

	f.excludePatterns = make([]*regexp.Regexp, 0, len(allPatterns))
	for _, pattern := range allPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			f.excludePatterns = append(f.excludePatterns, re)
		}
	}

	return f
}

func (f *Filter) LoadGitignore(repoRoot string) error {
	if !f.respectGitignore {
		return nil
	}

	gitignorePath := filepath.Join(repoRoot, ".gitignore")
	file, err := os.Open(gitignorePath)
	if err != nil {
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		pattern := line
		if strings.HasPrefix(pattern, "!") {
			continue
		}

		if strings.HasSuffix(pattern, "/") {
			pattern = pattern + "**"
		}

		if strings.HasPrefix(pattern, "/") {
			pattern = strings.TrimPrefix(pattern, "/")
		} else {
			pattern = "**/" + pattern
		}

		if g, err := glob.Compile(pattern, '/'); err == nil {
			f.gitignoreGlobs = append(f.gitignoreGlobs, g)
		}
	}

	return scanner.Err()
}

func (f *Filter) ShouldInclude(path string) bool {
	path = filepath.ToSlash(path)

	for _, re := range f.excludePatterns {
		if re.MatchString(path) {
			return false
		}
	}

	if f.respectGitignore {
		for _, g := range f.gitignoreGlobs {
			if g.Match(path) {
				return false
			}
		}
	}

	ext := filepath.Ext(path)
	filename := filepath.Base(path)

	// Allow common files without extensions
	commonNoExtFiles := map[string]bool{
		"Makefile": true, "Dockerfile": true, "Rakefile": true,
		"Gemfile": true, "Procfile": true, "Vagrantfile": true,
	}

	if ext == "" {
		return commonNoExtFiles[filename]
	}

	if !f.allowedExts[ext] {
		return false
	}

	return true
}

func CountLines(filePath string) (int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	const bufferSize = 32 * 1024
	buf := make([]byte, bufferSize)
	count := 0
	firstChunk := true

	for {
		n, err := file.Read(buf)
		if n > 0 {
			if firstChunk {
				if bytes.IndexByte(buf[:n], 0) != -1 {
					return 0, nil
				}
				firstChunk = false
			}
			count += bytes.Count(buf[:n], []byte{'\n'})
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return count, err
		}
	}

	return count, nil
}
