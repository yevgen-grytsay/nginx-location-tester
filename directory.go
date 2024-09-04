package main

import (
	"log"
	"os"
	"path/filepath"
)

func CollectRelativeFilePaths(basePath string) []string {
	var result []string
	visitor := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(basePath, path)
		result = append(result, relPath)

		return nil
	}
	err := filepath.Walk(basePath, visitor)
	if err != nil {
		log.Println(err)
	}

	return result
}
