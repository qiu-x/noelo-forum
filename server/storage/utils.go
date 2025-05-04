package storage

import (
	"fmt"
	"os"
)

type FilePath string
type DirPath string

func createPaths(entries map[any]string) error {
	for path, content := range entries {
		switch p := path.(type) {
		case DirPath:
			if err := os.Mkdir(string(p), 0755); err != nil {
				return fmt.Errorf("failed to create dir %s: %w", p, err)
			}
		case FilePath:
			f, err := os.Create(string(p))
			if err != nil {
				return fmt.Errorf("failed to create file %s: %w", p, err)
			}
			if _, err := f.WriteString(content); err != nil {
				_ = f.Close()
				return fmt.Errorf("failed to write to file %s: %w", p, err)
			}
			if err := f.Close(); err != nil {
				return fmt.Errorf("failed to close file %s: %w", p, err)
			}
		default:
			return fmt.Errorf("unsupported path type %T", path)
		}
	}
	return nil
}
