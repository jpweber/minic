package main

import (
	"io/ioutil"
	"log"
)

func listLocalFiles(dir string) map[string]bool {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal("Error reading local directory", err)
	}

	localFiles := make(map[string]bool)
	for _, f := range files {
		localFiles[f.Name()] = true
	}

	return localFiles
}

func newFiles(remoteFiles, localFiles map[string]bool) []string {
	filesToDownload := []string{}
	for k := range remoteFiles {
		if localFiles[k] == false {
			filesToDownload = append(filesToDownload, k)
		}
	}
	return filesToDownload
}
