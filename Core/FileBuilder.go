package Core

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func MakeDir(directory string) {
	fullDirPath, err := filepath.Abs("../Work/" + directory)
	if err != nil {
		log.Fatal(err)
	}
	errx := os.Mkdir(fullDirPath, os.ModeDir)
	if errx != nil {
		log.Fatal(errx)
	}
}

func MakeFile(filename string, directory string, size int64) error {
	fullFilePath := filepath.Join(directory, filename)
	f, err := os.Create(fullFilePath)
	if err != nil {
		return err
	}
	if err := f.Truncate(size); err != nil {
		return err
	}
	return nil
}

func CreateTestFiles(config TestConfig, totalRequests int) error {
	baseDir := filepath.Join("Work", "testfiles", config.TestID)
	os.MkdirAll(baseDir, 0755)

	bar := NewProgressBar(totalRequests)
	defer bar.Finish()

	fileCounter := 0
	var generatedFiles []string

	for i := range config.FilesizePolicies {
		policy := &config.FilesizePolicies[i]
		policy.Count = int(float64(totalRequests) * float64(policy.Percent) / 100)
		if policy.Count < 1 {
			policy.Count = 1
		}

		sizeKB := policy.Size
		unit := policy.Unit
		if policy.Unit == "MB" {
			sizeKB *= 1024
		} else if policy.Unit == "GB" {
			sizeKB *= 1024 * 1024
		}

		for j := 0; j < policy.Count; j++ {
			filename := fmt.Sprintf("%d%s_%d.dat", policy.Size, unit, j+1)
			if err := MakeFile(filename, baseDir, int64(sizeKB*1024)); err != nil {
				return err
			}
			fileCounter++
			bar.Update(fileCounter)
			generatedFiles = append(generatedFiles, filename)
		}
	}

	// Write file list to manifest
	manifestPath := filepath.Join(baseDir, "files.manifest")
	if err := os.WriteFile(manifestPath, []byte(strings.Join(generatedFiles, "\n")), 0644); err != nil {
		return fmt.Errorf("failed to write manifest: %w", err)
	}

	log.Printf("Successfully generated %d files in directory: %s", fileCounter, baseDir)
	log.Printf("File size distribution:")
	for _, p := range config.FilesizePolicies {
		log.Printf(" - %d%s: %d files", p.Size, p.Unit, p.Count)
	}
	return nil
}

func getFileList(testID string) []string {
	manifestPath := filepath.Join("Work", "testfiles", testID, "files.manifest")
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Printf("Error reading manifest: %v", err)
		return nil
	}
	return strings.Split(string(data), "\n")
}
