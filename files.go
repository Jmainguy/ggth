package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func copyFiles(src, dst string) {
	srcFiles := getAllFiles(src)
	for _, file := range srcFiles {
		destFile := strings.TrimPrefix(file, src)
		destFile = dst + "/" + destFile
		bytesRead, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		err = ioutil.WriteFile(destFile, bytesRead, 0644)

		if err != nil {
			log.Fatal(err)
		}

	}
}

// https://golang.cafe/blog/how-to-list-files-in-a-directory-in-go.html
func getAllFiles(dirPath string) (files []string) {
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}

		if !info.IsDir() {
			// Don't list .git/ files
			if !strings.Contains(path, ".git/") && !strings.Contains(path, ".template/") {
				files = append(files, path)
			}
		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}
	return files
}
