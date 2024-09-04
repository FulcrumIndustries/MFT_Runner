package main

import (
	"log"
	"os"
	"path/filepath"
)

func makeDir(directory string) {
	fullDirPath, err := filepath.Abs("../Work/" + directory)
	if err != nil {
		log.Fatal(err)
	}
	errx := os.Mkdir(fullDirPath, os.ModeDir)
	if errx != nil {
		log.Fatal(errx)
	}
}
func makeFile(filename string, directory string, size int64) {
	fullFilePath, err := filepath.Abs("../Work/" + directory + "/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(fullFilePath)
	if err != nil {
		log.Fatal(err)
	}
	if err := f.Truncate(size); err != nil {
		log.Fatal(err)
	}
}
