package main

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"time"
)

func main() {
	mux := http.NewServeMux()

	mux.Handle("/app/", http.StripPrefix("/app", http.FileServer(http.Dir("./static/"))))
	mux.HandleFunc("/healthz", handleReadyPath)

	server := &http.Server{
		Handler: mux,
		Addr:    ":8080",
	}

	err := server.ListenAndServe()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func handleReadyPath(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "text/plain")
	writer.Write([]byte("OK"))
	http.ServeContent(writer, request, "", time.Now(), &bytes.Reader{})
}
