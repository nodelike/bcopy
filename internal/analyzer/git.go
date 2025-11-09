package analyzer

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

func IsGitRepo(path string) bool {
	_, err := findGitRoot(path)
	return err == nil
}

func GetRepoRoot(path string) (string, error) {
	return findGitRoot(path)
}

// findGitRoot walks up the directory tree to find the .git directory
func findGitRoot(startPath string) (string, error) {
	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	currentPath := absPath
	for {
		gitPath := filepath.Join(currentPath, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			// Check if .git is a directory or file (for submodules/worktrees)
			if info.IsDir() || info.Mode().IsRegular() {
				// Verify with go-git that it's a valid repo
				if _, err := git.PlainOpen(currentPath); err == nil {
					return currentPath, nil
				}
			}
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)

		// Check if we've reached the root
		if parentPath == currentPath {
			return "", os.ErrNotExist
		}

		currentPath = parentPath
	}
}
