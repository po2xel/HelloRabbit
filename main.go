package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"

	"./hub"
)

func main() {
	router := mux.NewRouter()

	hub := hub.New()
	go hub.Run()

	// Serve file under <site> folder.
	fs := http.FileServer(http.Dir("site/"))
	router.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.ServeWS(w, r)
	})
	router.PathPrefix("/").Handler(fs)
	router.Handle("/", fs)

	//log.Fatal(http.ListenAndServeTLS(":8080", "server.crt", "server.key", router))
	log.Fatal(http.ListenAndServe(":8080", router))
}
