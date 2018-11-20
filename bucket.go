package main

import (
	"log"

	minio "github.com/minio/minio-go"
)

// Checking if a bucket exists
func bucketExists(client *minio.Client, bucketName string) {
	found, err := client.BucketExists(bucketName)
	if err != nil {
		log.Fatalln(err)
	}

	if found {
		log.Println(bucketName, "Bucket found.")
	} else {
		log.Fatalln("Bucket not found.", bucketName)

	}
}

// listing objects in a bucket
func listObjects(client *minio.Client, bucketName string) (map[string]bool, error) {

	// Create a done channel to control 'ListObjectsV2' go routine.
	doneCh := make(chan struct{})

	// Indicate to our routine to exit cleanly upon return.
	defer close(doneCh)

	isRecursive := true
	objectCh := client.ListObjectsV2(bucketName, "", isRecursive, doneCh)
	remoteFiles := make(map[string]bool)
	for object := range objectCh {
		if object.Err != nil {
			return remoteFiles, object.Err
		}
		remoteFiles[object.Key] = true
	}
	return remoteFiles, nil
}
