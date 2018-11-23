package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"

	minio "github.com/minio/minio-go"
)

func getEtag(path string, partSizeMb int) string {
	partSize := partSizeMb * 1024 * 1024
	content, _ := ioutil.ReadFile(path)
	fileSize := len(content)
	contentToHash := content
	parts := 0

	if fileSize > partSize {
		pos := 0
		contentToHash = make([]byte, 0)
		for fileSize > pos {
			endpos := pos + partSize
			if endpos >= fileSize {
				endpos = fileSize
			}
			hash := md5.Sum(content[pos:endpos])
			contentToHash = append(contentToHash, hash[:]...)
			pos += partSize
			parts++
		}
	}

	hash := md5.Sum(contentToHash)
	etag := fmt.Sprintf("%x", hash)
	if parts > 0 {
		etag += fmt.Sprintf("-%d", parts)
	}
	return etag
}

// checksumMatch - returns true if they match, false if they do not match
func checksumMatch(remoteChecksum, localFilePath string) bool {

	// check if local file exists. If not return false right away
	_, err := os.Stat(localFilePath)
	if err != nil {
		log.Println("Could not find file", localFilePath, "Skipping checksum and downloading")
		return false
	}

	// get md5sum of local file
	localChecksum := getEtag(localFilePath, 64)
	if err != nil {
		log.Println("Error getting local md5 checksum", err)
		return false
	}

	log.Println("Local Sum:", localChecksum, "Remote Sum:", remoteChecksum)
	// compare local md5 with remote md5
	if string(localChecksum) == remoteChecksum {
		return true
	}
	// return false by default
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
		go func(wg *sync.WaitGroup, fileName string) {
			stat, err := client.StatObject(bucketName, fileName, minio.StatObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}

			// if the checksums match we don't need to download because they are
			// the same thing
			if checksumMatch(stat.ETag, dest+"/"+fileName) {
				log.Printf("%s/%s Has Not Changed. Not downloading.", dest, fileName)
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
