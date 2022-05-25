package main

import (
	"fmt"
	"github.com/tus/tusd/pkg/filestore"
	tusd "github.com/tus/tusd/pkg/handler"
	"log"
	"net/http"
	"os"
)

func main() {
	err := realMain()
	if err != nil {
		log.Fatalln("Error:", err)
	}

}

func PreUploadHook(event tusd.HookEvent) error {
	fmt.Println("PreUploadHook invoked")
	return nil
}

func PreFinishResponseHook(event tusd.HookEvent) error {
	fmt.Println("PostUploadHook invoked")
	fmt.Println("Dump: %+v", event)
	filename, ok := event.Upload.MetaData["filename"]
	if !ok {
		return fmt.Errorf("no filename in metadata")
	}
	storage, ok := event.Upload.Storage["Path"]
	if !ok {
		return fmt.Errorf("no storage in metadata")
	}
	log.Printf("Renameing file %s to %s", storage, filename)
	err := os.Rename(storage, filename)
	if err != nil {
		return fmt.Errorf("error renaming file: %w", err)
	}
	return nil
}

func realMain() error {
	logger := log.New(os.Stdout, "tusd ", 0)
	store := filestore.FileStore{
		Path: "./uploads",
	}
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)

	tusHandler, err := tusd.NewHandler(tusd.Config{
		BasePath:                  "/files/",
		StoreComposer:             composer,
		NotifyCompleteUploads:     true,
		NotifyTerminatedUploads:   true,
		Logger:                    logger,
		PreUploadCreateCallback:   PreUploadHook,
		PreFinishResponseCallback: PreFinishResponseHook,
	})

	if err != nil {
		panic(fmt.Errorf("unable to create tusHandler: %s", err))
	}

	// Start another goroutine for receiving events from the tusHandler whenever
	// an upload is completed or terminated. The event will contain details about the upload
	// itself and the relevant HTTP request.
	go func() {
		for {
			select {
			case event := <-tusHandler.CompleteUploads:
				fmt.Printf("Upload %s finished\n", event.Upload.ID)
			case event := <-tusHandler.TerminatedUploads:
				fmt.Printf("Upload %s terminated\n", event.Upload.ID)
			}
		}
	}()
	http.Handle("/files/", http.StripPrefix("/files/", tusHandler))

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "8080"
	}
	log.Println("Listening on port", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%s", port), nil); err != nil {
		log.Fatal(err)
	}
	return nil
}

func indexHandler(writer http.ResponseWriter, request *http.Request) {
	fmt.Fprintf(writer, "Hello, world!\n")
}
