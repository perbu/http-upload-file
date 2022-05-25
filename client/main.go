package main

import (
	"fmt"
	"github.com/eventials/go-tus"
	"log"
	"os"
)

func main() {
	err := realMain()
	if err != nil {
		log.Fatal(err)
	}
}

func realMain() error {
	f, err := os.Open("/tmp/bigfile")

	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}

	defer f.Close()

	// create the tus client.
	client, err := tus.NewClient("http://localhost:8080/files/", nil)
	if err != nil {
		return fmt.Errorf("failed to create tus client: %w", err)
	}

	// create an upload from a file.
	upload, err := tus.NewUploadFromFile(f)
	if err != nil {
		return fmt.Errorf("failed to create upload from file: %w", err)
	}

	// create the uploader.
	uploader, err := client.CreateUpload(upload)
	if err != nil {
		return fmt.Errorf("failed to create uploader: %w", err)
	}

	// start the uploading process.
	err = uploader.Upload()
	if err != nil {
		return fmt.Errorf("failed to upload: %w", err)
	}

	return nil
}
