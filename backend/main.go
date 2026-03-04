package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"schedulehelper/cron"
	"schedulehelper/db"
	"schedulehelper/handlers"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "OK")
}

func serveFrontend(w http.ResponseWriter, r *http.Request) {
	path := filepath.Join("frontend", "dist", filepath.Clean(r.URL.Path))
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.ServeFile(w, r, filepath.Join("frontend", "dist", "index.html"))
		return
	}
	http.ServeFile(w, r, path)
}

func main() {
	if err := db.InitDB("schedulehelper.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	cron.StartCronJobs()

	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/users", handlers.GetUsersHandler)
	http.HandleFunc("/api/register", handlers.RegisterHandler)
	http.HandleFunc("/api/login", handlers.LoginHandler)
	http.HandleFunc("/api/profile", handlers.ProfileHandler)
	http.HandleFunc("/api/slots", handlers.SlotsHandler)
	http.HandleFunc("/api/events", handlers.EventsHandler)
	http.HandleFunc("/api/calendar", handlers.GetCalendarDataHandler)
	http.HandleFunc("/api/sse", handlers.SSEHandler)
	
	http.HandleFunc("/", serveFrontend)

	log.Println("Server listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}