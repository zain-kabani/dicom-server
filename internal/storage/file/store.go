package file

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

type Store struct {
	stagingDir string
	finalDir   string
}

func NewStore(stagingDir, finalDir string) *Store {
	// Create directories if they don't exist
	dirs := []string{stagingDir, finalDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			fmt.Printf("Failed to create directory %s: %v\n", dir, err)
			os.Exit(1)
		}
	}

	return &Store{
		stagingDir: stagingDir,
		finalDir:   finalDir,
	}
}

func (s *Store) SaveToStaging(content io.Reader) (string, error) {
	stagingName := uuid.New().String()
	stagingPath := filepath.Join(s.stagingDir, stagingName)

	dst, err := os.Create(stagingPath)
	if err != nil {
		return "", fmt.Errorf("failed to create staging file: %w", err)
	}
	defer dst.Close()

	if _, err := io.Copy(dst, content); err != nil {
		os.Remove(stagingPath) // cleanup on error
		return "", fmt.Errorf("failed to write staging file: %w", err)
	}

	return stagingPath, nil
}

func (s *Store) CreateFinalDirectory(hash string) (string, error) {
	dirPath := filepath.Join(s.finalDir, hash)
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	return dirPath, nil
}

func (s *Store) MoveDICOM(sourcePath, hash string) (string, error) {
	dirPath := filepath.Join(s.finalDir, hash)
	finalPath := filepath.Join(dirPath, "original.dcm")

	if err := os.Rename(sourcePath, finalPath); err != nil {
		return "", fmt.Errorf("failed to move DICOM file: %w", err)
	}
	return finalPath, nil
}

func (s *Store) SavePNG(imageData []byte, hash string) (string, error) {
	dirPath := filepath.Join(s.finalDir, hash)
	pngPath := filepath.Join(dirPath, "preview.png")

	if err := os.WriteFile(pngPath, imageData, 0644); err != nil {
		return "", fmt.Errorf("failed to save PNG: %w", err)
	}
	return pngPath, nil
}

func (s *Store) Cleanup(path string) error {
	if path == "" {
		return nil
	}
	return os.RemoveAll(path)
}

func (s *Store) CleanupDirectory(hash string) error {
	return os.RemoveAll(filepath.Join(s.finalDir, hash))
}

func (s *Store) FinalDir() string {
	return s.finalDir
}
