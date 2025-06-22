package utils

import (
	"os"
	"path/filepath"
	"strings"
)

func GatherManifestFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".json") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
