package main

import (
	"context"
	"log"
	"time"

	minio "github.com/minio/minio-go"
)

func getFiles(client *minio.Client, filesToDownload []string, bucketName, dest string) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	for _, fileName := range filesToDownload {
		// DEBUG
		log.Println("downloading", fileName)
		err := client.FGetObjectWithContext(ctx, bucketName, fileName, dest+"/"+fileName, minio.GetObjectOptions{})
		if err != nil {
			log.Println("Error downloading file ", err)
			return
		}
	}
}
