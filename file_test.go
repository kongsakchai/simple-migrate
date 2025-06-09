package migrate

import (
	"slices"
	"testing"
)

func TestFindMigrationFiles(t *testing.T) {
	t.Run("should return files with suffix 'up.sql'", func(t *testing.T) {
		dir := "./example"

		files, err := findMigrationFiles(dir, SuffixUp, false)
		if err != nil {
			t.Fatalf("Failed to find files: %v", err)
		}

		expectedFiles := []string{"001_example.up.sql", "002_example.up.sql"}
		if len(files) != len(expectedFiles) {
			t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(files))
		}

		if !slices.Equal(files, expectedFiles) {
			t.Fatalf("Expected files %v, got %v", expectedFiles, files)
		}
	})

	t.Run("should return files with suffix 'up.sql' in inverse order", func(t *testing.T) {
		dir := "./example"

		files, err := findMigrationFiles(dir, SuffixUp, true)
		if err != nil {
			t.Fatalf("Failed to find files: %v", err)
		}

		expectedFiles := []string{"002_example.up.sql", "001_example.up.sql"}
		if len(files) != len(expectedFiles) {
			t.Fatalf("Expected %d files, got %d", len(expectedFiles), len(files))
		}

		if !slices.Equal(files, expectedFiles) {
			t.Fatalf("Expected files %v, got %v", expectedFiles, files)
		}
	})

	t.Run("should return error when directory does not exist", func(t *testing.T) {
		dir := "./nonexistent"

		_, err := findMigrationFiles(dir, SuffixUp, false)
		if err == nil {
			t.Fatal("Expected an error when directory does not exist, but got none")
		}
	})
}

func TestReadMigrationFile(t *testing.T) {
	t.Run("should read migration file content", func(t *testing.T) {
		dir := "./example"
		fileName := "example.txt"

		content, err := readMigrationFile(dir, fileName)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		expectedContent := "Example\n"
		if string(content) != expectedContent {
			t.Fatalf("Expected content %q, got %q", expectedContent, content)
		}
	})

	t.Run("should return error for non-existent file", func(t *testing.T) {
		dir := "./example"
		fileName := "nonexistent.sql"

		_, err := readMigrationFile(dir, fileName)
		if err == nil {
			t.Fatal("Expected an error when reading a non-existent file, but got none")
		}
	})
}
