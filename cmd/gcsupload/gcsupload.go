package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
)

func main() {

}

// uploadFile uploads an object.
func uploadFile(w io.Writer, bucket, objectName string, fileBytes []byte) error {
	// bucket := "bucket-name"
	// objectName := "objectName-name"
	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("storage.NewClient: %v", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	// Upload an object with storage.Writer.
	wc := client.Bucket(bucket).Object(objectName).NewWriter(ctx)
	if _, err = io.Copy(wc, bytes.NewBuffer(fileBytes)); err != nil {
		return fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("Writer.Close: %v", err)
	}
	fmt.Fprintf(w, "Blob %v uploaded.\n", objectName)
	return nil
}

