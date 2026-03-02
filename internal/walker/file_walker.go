package walker

import (
	"os"
	"path/filepath"
)

// FileWalkerImpl implements the FileWalker interface
type FileWalkerImpl struct {
	extensions    []string
	defaultIgnore []string
}

// NewFileWalker creates a new FileWalker instance
func NewFileWalker() *FileWalkerImpl {
	return &FileWalkerImpl{
		extensions:    []string{".js", ".ts", ".go", ".py"},
		defaultIgnore: []string{".git", "node_modules", "vendor"},
	}
}

// Walk recursively traverses directories and returns file paths matching the criteria
func (w *FileWalkerImpl) Walk(rootDir string, ignoreDirs []string) ([]string, error) {
	var files []string

	// Combine default ignore list with custom ignore list
	ignoreMap := make(map[string]bool)
	for _, dir := range w.defaultIgnore {
		ignoreMap[dir] = true
	}
	for _, dir := range ignoreDirs {
		ignoreMap[dir] = true
	}

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip symbolic links
		if info.Mode()&os.ModeSymlink != 0 {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if directory should be ignored
		if info.IsDir() {
			dirName := info.Name()
			if ignoreMap[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		// Check if file has a matching extension (case-sensitive)
		ext := filepath.Ext(path)
		for _, validExt := range w.extensions {
			if ext == validExt {
				files = append(files, path)
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return files, nil
}
