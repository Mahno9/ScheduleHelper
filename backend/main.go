package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/robfig/cron/v3"
	"schedulehelper/backend/db"
	"schedulehelper/backend/handlers"
	"schedulehelper/backend/services"
)

func main() {
	dsn := os.Getenv("DB_PATH")
	if dsn == "" {
		dsn = "./schedulehelper.db"
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	staticDir := os.Getenv("STATIC_DIR")
	if staticDir == "" {
		staticDir = "./static"
	}

	// Open DB
	database, err := db.Open(dsn)
	if err != nil {
		log.Fatalf("failed to open database %s: %v", dsn, err)
	}
	defer database.Close()

	// SSE hub
	hub := services.NewSSEHub()

	// Handler
	h := handlers.New(database, hub)

	// Router
	r := mux.NewRouter()

	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", h.Health).Methods(http.MethodGet)
	api.HandleFunc("/users", h.GetUsers).Methods(http.MethodGet)
	api.HandleFunc("/users/sort", h.UpdateSortOrder).Methods(http.MethodPut)
	api.HandleFunc("/register", h.Register).Methods(http.MethodPost)
	api.HandleFunc("/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/profile/{id:[0-9]+}", h.GetProfile).Methods(http.MethodGet)
	api.HandleFunc("/profile/{id:[0-9]+}", h.UpdateProfile).Methods(http.MethodPut)
	api.HandleFunc("/profile/{id:[0-9]+}", h.DeleteProfile).Methods(http.MethodDelete)
	api.HandleFunc("/slots", h.GetSlots).Methods(http.MethodGet)
	api.HandleFunc("/slots", h.CreateSlot).Methods(http.MethodPost)
	api.HandleFunc("/slots/{id:[0-9]+}", h.UpdateSlot).Methods(http.MethodPut)
	api.HandleFunc("/slots/{id:[0-9]+}", h.DeleteSlot).Methods(http.MethodDelete)
	api.HandleFunc("/events", h.GetEvents).Methods(http.MethodGet)
	api.HandleFunc("/events", h.CreateEvent).Methods(http.MethodPost)
	api.HandleFunc("/events/{id:[0-9]+}", h.UpdateEvent).Methods(http.MethodPut)
	api.HandleFunc("/events/{id:[0-9]+}", h.DeleteEvent).Methods(http.MethodDelete)
	api.HandleFunc("/calendar", h.GetCalendar).Methods(http.MethodGet)
	api.HandleFunc("/events/sse", h.SSEEvents).Methods(http.MethodGet)

	// CORS middleware for development
	r.Use(corsMiddleware)

	// Serve static frontend files
	r.PathPrefix("/").Handler(spaHandler(staticDir))

	// Cron: cleanup old data weekly
	c := cron.New()
	c.AddFunc("@weekly", func() {
		log.Println("Running scheduled cleanup...")
		database.CleanupOldData()
	})
	c.Start()
	defer c.Stop()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 0, // 0 for SSE streaming
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("ScheduleHelper server starting on :%s (DB: %s, static: %s)", port, dsn, staticDir)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves a Single Page Application: if a file doesn't exist, serve index.html
type spaHandlerType struct {
	dir string
	fs  http.Handler
}

func spaHandler(dir string) http.Handler {
	return &spaHandlerType{
		dir: dir,
		fs:  http.FileServer(http.Dir(dir)),
	}
}

func (s *spaHandlerType) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := s.dir + r.URL.Path
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		http.ServeFile(w, r, s.dir+"/index.html")
		return
	}
	s.fs.ServeHTTP(w, r)
}
