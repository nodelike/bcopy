package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func ValidatePath(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	cleanPath := filepath.Clean(absPath)

	if cleanPath == "/" {
		return fmt.Errorf("refusing to run in root directory (/). This could scan your entire system")
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		cleanHome := filepath.Clean(homeDir)
		if cleanPath == cleanHome {
			return fmt.Errorf("refusing to run in home directory (%s). Please run in a specific project directory", cleanHome)
		}
	}

	dangerousDirs := []string{
		"/usr",
		"/etc",
		"/var",
		"/bin",
		"/sbin",
		"/boot",
		"/sys",
		"/proc",
		"/dev",
		"/System",
		"/Library",
		"/Applications",
		"/Volumes",
		"/private",
		"/opt",
		"/root",
		"/tmp",
		"/Windows",
		"/Program Files",
		"/Program Files (x86)",
	}

	for _, dangerousDir := range dangerousDirs {
		cleanDangerous := filepath.Clean(dangerousDir)
		if cleanPath == cleanDangerous {
			return fmt.Errorf("refusing to run in system directory (%s). This is a protected system location", cleanPath)
		}
	}

	pathDepth := strings.Count(cleanPath, string(os.PathSeparator))
	if pathDepth <= 2 && cleanPath != "/" {
		parentDir := filepath.Dir(cleanPath)
		if parentDir == "/" || (homeDir != "" && parentDir == filepath.Dir(homeDir)) {
			return fmt.Errorf("refusing to run at (%s). This directory is too broad. Please run in a specific project directory", cleanPath)
		}
	}

	return nil
}

func ShouldWarnLargeDirectory(path string) (bool, string) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, ""
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return false, ""
	}

	if strings.HasPrefix(absPath, homeDir) {
		relPath, err := filepath.Rel(homeDir, absPath)
		if err == nil && !strings.Contains(relPath, string(os.PathSeparator)) {
			return true, fmt.Sprintf("Warning: Analyzing a top-level directory in your home folder (%s). This may take a while.", filepath.Base(absPath))
		}
	}

	return false, ""
}
