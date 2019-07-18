package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

func runServer(config Config) {
	router := mux.NewRouter()
	router.HandleFunc("/", Index)
	router.HandleFunc("/team/{key}", TeamReq)
	router.HandleFunc("/reset", config.ResetReq)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome!")
}

func TeamReq(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Category: %v\n", vars["key"])
}

func (config *Config) ResetReq(w http.ResponseWriter, r *http.Request) {
	reset(config)
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "Done")
}
