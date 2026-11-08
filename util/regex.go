package util

import (
	"fmt"
	"log/slog"
	"os"
	"regexp"
)

func RegexMatchFilePaths(dir string, recursive bool, regex string, logger *slog.Logger) (paths []string, err error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	paths = make([]string, 0)

	validID := regexp.MustCompile(regex)
	for _, entry := range entries {
		if entry.IsDir() {
			if !recursive {
				continue
			}
			recursiveMatches, err := RegexMatchFilePaths(entry.Name(), true, regex, logger)
			if err != nil {
				return nil, fmt.Errorf("recursive search failed: %w", err)
			}
			paths = append(paths, recursiveMatches...)
		}

		if validID.MatchString(entry.Name()) {
			paths = append(paths, entry.Name())
		}
	}

	return paths, nil
}

func RegexMatchDirEntries(dir string, recursive bool, regex string, logger *slog.Logger) (entries []os.DirEntry, err error) {
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("could not read directory: %w", err)
	}

	entries = make([]os.DirEntry, 0)

	validID := regexp.MustCompile(regex)
	for _, entry := range dirEntries {
		if entry.IsDir() {
			if !recursive {
				continue
			}
			recursiveMatches, err := RegexMatchDirEntries(entry.Name(), true, regex, logger)
			if err != nil {
				return nil, fmt.Errorf("recursive search failed: %w", err)
			}
			entries = append(entries, recursiveMatches...)
		}

		if validID.MatchString(entry.Name()) {
			entries = append(entries, entry)
		}

	}

	return entries, nil
}

// if matchString() creates a matching string, the current value will be added to matches[]
func RegexMatchAny[T any](values []T, regex string, matchString func(idx int) string) (matches []T) {
	matches = make([]T, 0)

	validID := regexp.MustCompile(regex)
	for i, value := range values {
		if validID.MatchString(matchString(i)) {
			matches = append(matches, value)
		}
	}

	return matches
}
