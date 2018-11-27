package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"sync"

	minio "github.com/minio/minio-go"
)

func md5Hasher(filePath string) (string, error) {

	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string

	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
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

func eTagger(path string, partSizeMb int) string {
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

	// doing a normal md5checksum and my multi-part function checks.
	// There are instances where a regular md5 is used and where the multi-part
	// hash is used where you wouldn't always expect them. Checking both scenarios.
	//get md5sum of local file
	// fileContent, _ := ioutil.ReadFile(localFilePath)
	localMd5Checksum, err := md5Hasher(localFilePath)
	if err != nil {
		log.Println("Error getting local MD5 checksum", err)
		return false
	}

	// compare local md5 with remote md5
	if string(localMd5Checksum) == remoteChecksum {
		log.Println("Local MD5 CheckSum:", localMd5Checksum, "Remote MD5 CheckSum:", remoteChecksum)
		return true
	}

	// get multi-part ETag checksum of local file
	localEtagChecksum := eTagger(localFilePath, 64)
	if err != nil {
		log.Println("Error getting local Etag checksum", err)
		return false
	}

	// compare local multi-part etag with remote multi-part etag
	if string(localEtagChecksum) == remoteChecksum {
		log.Println("Local multi-part ETag:", localEtagChecksum, "Remote multi-part ETag:", remoteChecksum)
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
