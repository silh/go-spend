package main

import (
	"go-spend/log"
	"net/http"
)

func init() {
	log.Level = log.DebugLvl
}

func main() {
	mux := http.NewServeMux()
	mux.Handle("/users", http.HandlerFunc(handleCreateUser()))

	log.Info("Starting a server on port 8080...")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err.Error())
	}
}

func handleCreateUser() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Not supported", http.StatusBadRequest)
			return
		}
		// TODO
		w.WriteHeader(http.StatusCreated)
	}
}
