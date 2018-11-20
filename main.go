package main

import (
	"log"
	"os"
	"strings"

	"github.com/minio/minio-go"
)

func main() {

	minioURL := os.Getenv("MINIO_URL")
	accessKey := os.Getenv("ACCESSKEY")
	secretKey := os.Getenv("SECRETKEY")
	src := os.Getenv("SRC")
	dest := os.Getenv("DEST")

	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY and my-bucketname are
	// dummy values, please replace them with original values.

	// Requests are always secure (HTTPS) by default. Set secure=false to enable insecure (HTTP) access.
	// This boolean value is the last argument for New().

	// New returns an Amazon S3 compatible client object. API compatibility (v2 or v4) is automatically
	// determined based on the Endpoint value.
	// TODO: fetch this from env vars or config file
	minClient, err := minio.New(minioURL, accessKey, secretKey, false)
	if err != nil {
		log.Fatalln(err)
	}
	fileParts := strings.Split(src, "/")
	bucketName := fileParts[0]
	// first check if specified bucket exists
	bucketExists(minClient, bucketName)

	// determine if the user is trying to recursively download files in a dir
	// or a single file.

	// single file
	if len(fileParts) > 1 {
		fileName := fileParts[len(fileParts)-1]
		// converting file name to array so it is easy to reuse the same
		// fucntion for grabbing list of files
		filesToDownload := []string{fileName}
		getFiles(minClient, filesToDownload, bucketName, dest)
	}

	// files in a dir
	if len(fileParts) == 1 {
		// we are grabbing a while dir recursively
		// first get list of files
		remoteFiles, err := listObjects(minClient, bucketName)
		if err != nil {
			log.Fatalln("Error listing objects in ", bucketName, err)
		}

		// check if files already exist in destination, if not download them
		// localFiles := listLocalFiles(dest)

		// now determine files we have to actually download
		// and which ones we can skip
		// filesToDownload := newFiles(remoteFiles, localFiles)

		// go get them files
		getFiles(minClient, remoteFiles, bucketName, dest)

	}

}
