package main

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"log"
	"os"
	"sync"

	minio "github.com/minio/minio-go"
)

func hasher(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	//Open the passed argument and check for any error
	log.Println("Opening", filePath, "to hash contents")
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
		log.Println("Could not find file", localFilePath, "locally to checksum")
		return false
	}

	// get md5sum of local file
	localChecksum, err := hasher(localFilePath)
	if err != nil {
		log.Println("Error getting local md5 checksum", err)
		return false
	}

	// DEBUG
	log.Println("Local Sum:", localChecksum, "Remote Sum:", remoteChecksum)
	// compare local md5 with remote md5
	if localChecksum == remoteChecksum {
		return true
	}
	return false
}

func getFiles(client *minio.Client, filesToDownload []string, bucketName, dest string) {

	concurrency := 2
	workerPool := make(chan bool, concurrency)
	for i := 0; i < concurrency; i++ {
		workerPool <- true
	}

	var wg sync.WaitGroup
	wg.Add(len(filesToDownload))

	for _, fileName := range filesToDownload {
		<-workerPool
		// DEBUG
		// log.Println("Stat", bucketName, fileName)
		go func(wg *sync.WaitGroup, fileName string) {
			stat, err := client.StatObject(bucketName, fileName, minio.StatObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}

			// if the checksums match we don't need to download because they are
			// the same thing
			log.Println("Etag from", bucketName, fileName)
			if checksumMatch(stat.ETag, dest+"/"+fileName) {
				workerPool <- true
				wg.Done()
				return
			}

			log.Println("downloading", dest+"/"+fileName)
			// Spin this out in to go routines so we can download concurrently.
			// might want to buffer this with channels though so we don't saturate things
			err = client.FGetObject(bucketName, fileName, dest+"/"+fileName, minio.GetObjectOptions{})
			if err != nil {
				log.Println("Error downloading file ", err)
				return
			}

			workerPool <- true
			wg.Done()
		}(&wg, fileName)

	}
	wg.Wait()
}
