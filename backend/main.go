package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("frontend", "dist", filepath.Clean(r.URL.Path))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// SPA routing fallback
		http.ServeFile(w, r, filepath.Join("frontend", "dist", "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}

func main() {
	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/", serveFrontend)

	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}