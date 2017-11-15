package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	// Serve file under <site> folder.
	fs := http.FileServer(http.Dir("site/"))
	router.PathPrefix("/").Handler(fs)
	router.Handle("/", fs)

	log.Fatal(http.ListenAndServeTLS(":8080", "server.crt", "server.key", router))
	// log.Fatal(http.ListenAndServe(":8080", router))
}
