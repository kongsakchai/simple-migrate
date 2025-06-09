package migrate

import (
	"os"
	"path"
	"slices"
	"strings"
)

func findMigrationFiles(dir, suffix string, invert bool) ([]string, error) {
	entry, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, file := range entry {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), suffix) {
			files = append(files, file.Name())
		}
	}
	slices.SortFunc(files, func(a, b string) int {
		if !invert {
			return strings.Compare(a, b)
		}
		return strings.Compare(b, a)
	})

	return files, nil
}

func readMigrationFile(dir, fileName string) ([]byte, error) {
	b, err := os.ReadFile(path.Join(dir, fileName))
	if err != nil {
		return nil, err
	}

	return b, nil
}
