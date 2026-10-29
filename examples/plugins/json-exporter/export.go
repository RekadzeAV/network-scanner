package main

import (
	"fmt"
	"os"
)

// writeToFile пишет строку data в файл path (перезаписывает).
func writeToFile(path, data string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("create export file: %w", err)
	}

	if _, err := f.WriteString(data); err != nil {
		_ = f.Close()
		return fmt.Errorf("write export file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close export file: %w", err)
	}
	return nil
}
