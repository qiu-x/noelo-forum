package storage

import (
	"fmt"
	"os"
)

func createPaths(dirs []string, files map[string]string) error {
	if err := createDirs(dirs); err != nil {
		return err
	}
	if err := writeFiles(files); err != nil {
		return err
	}
	return nil
}

func createDirs(entries []string) error {
	if entries == nil {
		return nil
	}

	for _, path := range entries {
		if err := os.Mkdir(path, 0750); err != nil {
			return fmt.Errorf("failed to create dir %s: %w", path, err)
		}
	}
	return nil
}

func writeFiles(entries map[string]string) error {
	if entries == nil {
		return nil
	}

	for path, content := range entries {
		f, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", path, err)
		}
		if _, err := f.WriteString(content); err != nil {
			_ = f.Close()
			return fmt.Errorf("failed to write to file %s: %w", path, err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("failed to close file %s: %w", path, err)
		}
	}
	return nil
}
