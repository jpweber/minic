package main

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"os"
	"sync"

	minio "github.com/minio/minio-go"
)

func fileOpener(filePath string) (*os.File, error) {
	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	//Tell the program to call the following function when the current function returns
	// 	defer file.Close()
	// we can't actually defer close here we need to hold it open and close it later when we are done chunking

}

func chunker(chunkSize, offSet int64, file *os.File) ([]byte, error) {
	// create byte array to store our chunk of bytes
	byteChunk := make([]byte, chunkSize)

	// seek to the correct position of the file
	// When needing to hash mult-part files each part is hashed on its own
	// we have to seek to the right part of the file to read in just those bytes
	// as if we were uploading the file
	_, err := file.Seek(offSet, 0)
	if err != nil {
		return nil, err
	}

	// read in our chunk of data
	byteNum, err := file.Read(byteChunk)
	if err != nil {
		return nil, err
	}

	return byteChunk, nil
}

func hasher(dataChunk []byte) string {

	//Initialize variable returnMD5String now in case an error has to be returned
	var md5String string

	//Open a new hash interface to write to
	hash := md5.New()

	//Get the 16 bytes hash
	hashInBytes := hash.Sum(dataChunk)[:16]

	//Convert the bytes to a string
	md5String = hex.EncodeToString(hashInBytes)

	return md5String

}


// checksumMatch - returns true if they match, false if they do not match
func checksumMatch(remoteChecksum, localFilePath string, numParts int) bool {

	// check if local file exists. If not return false right away
	_, err := os.Stat(localFilePath)
	if err != nil {
		log.Println("Could not find file", localFilePath, "locally to checksum")
		return false
	}

	// get md5sum of local file
	file, err := fileOpener(localFilePath)
	if err != nil {
		log.Println("Error opening file", err)
		return false
	}
	defer file.Close()

	// init var to pass to hasher
	localBytes := []byte{}
	if numParts == 1 {
		//the hash is of the single file, not the combined chunks
		fileInfo, err := file.Stat()
		if err != nil {
			log.Println("Error stating local file", localFilePath, err)
			return false
		}
		localBytes = chunker(fi.Size,0,file)
		
	}else{
		// we are dealing with a hash that is comprised of multiple chunks
		chunkHashes := []string
		
		// assuming 64mb file size for now
		chunkSize := 64000
		offSet := 0
		for i := 0 ; i < numParts ; i++ {
			chunkBytes := chunker(chunkSize, offSet, file)
			chunkHash := hasher(chunkBytes)
			chunkHashes = append(chunkHashes, chunkHash)
			offset += chunkSize
		}
		
		// DEBUG
		fmt.Println(chunkHashes)
		combinedChunkHashes := strings.Join(chunkHashes, "")
		localBytes = []byte(combinedChunkHashes)
		
	}
	
	localChecksum := hasher(localBytes)
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
		// DEBUG
		// log.Println("Stat", bucketName, fileName)
		go func(wg *sync.WaitGroup, fileName string) {
			stat, err := client.StatObject(bucketName, fileName, minio.StatObjectOptions{})
			if err != nil {
				log.Fatalln(err)
			}

			// read type from stat object to see if it is an octet stream
			// if it is we need to do the multi part hashing. The minio Etag format
			// is hash-numberofparts. Figure out the number of parts for use later
			multiPartCount := 1
			if stat.ContentType == "application/octet-stream" {
				multiPartCount = strings.Split(stat.Etag, "-")[1]
			}
			
			// if the checksums match we don't need to download because they are
			// the same thing
			log.Println("Etag from", bucketName, fileName)
			if checksumMatch(stat.ETag, dest+"/"+fileName, multiPartCount) {
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
