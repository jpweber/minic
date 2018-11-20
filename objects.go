package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/ioutil"
	"log"
	"os"
)

func hasher(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}

	//Tell the program to call the following function when the current function returns
	defer file.Close()

	//Open a new hash interface to write to
	hash := md5.New()

	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]

	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)

	return returnMD5String, nil

}

// checksumMatch - returns true if they match, false if they do not match
func checksumMatch(remoteChecksum, localFilePath string) bool {

	// check if local file exists. If not return false right away
	_, err := os.Stat(localFilePath)
	if err != nil {
		return false
	}

	// get md5sum of local file
	localChecksum, err := hasher(localFilePath)
	if err != nil {
		log.Println("Error getting local md5 checksum", err)
		return false
	}

	// DEBUG
	// fmt.Println("Local Sum:", localChecksum, "Remote Sum:", remoteChecksum)
	// compare local md5 with remote md5
	if localChecksum == remoteChecksum {
		return true
	}
	return false
}

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
