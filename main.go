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
	filePath := strings.Join(fileParts[1:], "/")

	// first check if specified bucket exists
	bucketExists(minClient, bucketName)

	// get list of remote file(s) to download
	remoteFiles, err := listObjects(minClient, bucketName, filePath)
	if err != nil {
		log.Fatalln("Error listing objects in ", bucketName, err)
	}

	// go get them files
	getFiles(minClient, remoteFiles, bucketName, dest)

}
