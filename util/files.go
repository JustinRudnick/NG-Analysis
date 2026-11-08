package util

import (
	"fmt"
	"log/slog"
	"os"
)

func OpenFiles(dir string, recursive bool, regex string, flags int, perm os.FileMode, logger *slog.Logger) (files []*os.File, err error) {
	entries, err := RegexMatchDirEntries(dir, recursive, regex, logger)
	if err != nil {
		LogError(logger, "Open Files: Could not resolve matching entries.")
		return nil, fmt.Errorf("could not resolve matching entries: %w", err)
	}

	files = make([]*os.File, 0)

	for _, entry := range entries {
		file, err := os.OpenFile(fmt.Sprintf("%s/%s", dir, entry.Name()), flags, perm)
		if err != nil && logger != nil {
			logger.Warn(fmt.Sprintf("File %s could not be opened", entry.Name()))
		}
		if err == nil {
			files = append(files, file)
		}
	}

	return files, nil
}

func CloseFiles(files []*os.File) {
	for _, file := range files {
		if file != nil {
			file.Close()
		}
	}
}

func remove[T any](slice []T, idx int) []T {
	return append(slice[:idx], slice[idx+1:]...)
}
