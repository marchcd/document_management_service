package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/marchcd/kai/internal/api/handlers"
	"github.com/marchcd/kai/internal/models"
	"github.com/marchcd/kai/internal/repository"
	"github.com/marchcd/kai/internal/service"
)

const sessionsFilePath = "static/sessions.json"

func main() {
	ctx := context.Background()
	rawConnString := os.Getenv("DB_URL")
	connString := os.ExpandEnv(rawConnString)
	if connString == "" {
		fmt.Fprintf(os.Stderr, "DB_URL env variable not set\n")
		os.Exit(1)
	}

	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		log.Fatalf("Unable to connect to db: %v", err)
	}
	defer dbpool.Close()

	// Read JSON file if exist (or create new empty one)
	var sessionsData []models.DirectionJSON
	file, err := os.Open(sessionsFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			sessionsData = []models.DirectionJSON{}
			emptyData, _ := json.MarshalIndent(sessionsData, "", " ")
			_ = os.WriteFile(sessionsFilePath, emptyData, 0o644)
		} else {
			log.Fatal("Cant open the file: ", err)
		}
	} else {
		defer file.Close()

		if err := json.NewDecoder(file).Decode(&sessionsData); err != nil && err != io.EOF {
			log.Fatal("JSON parsing error: ", err)
		}
	}
	//

	docRepo := repository.NewStore(dbpool)
	docService := service.NewDocumentService(docRepo, sessionsData)

	createHandler := handlers.NewCreatePDFHandler(docService)
	requestsHandler := handlers.NewRequestsHandler(docService)
	sessionsHandler := handlers.NewSessionsHandler(docService, sessionsFilePath)
	registryHandler := handlers.NewRegistryHandler(docService)
	statusHandler := handlers.NewStatusHandler(docService)
	loginHandler := handlers.NewLoginHandler()

	mainMux := http.NewServeMux()

	// Static pages
	mainMux.HandleFunc("GET /student", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/docData.html")
	})

	mainMux.Handle("/login", loginHandler)
	mainMux.HandleFunc("GET /logout", handlers.LogoutHandler)

	mainMux.HandleFunc("GET /requests", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/list.html")
	}))
	mainMux.HandleFunc("GET /sessions", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/sessions.html")
	}))
	mainMux.HandleFunc("GET /registry", handlers.AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/registry.html")
	}))
	mainMux.HandleFunc("GET /status", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/status.html")
	})

	// API
	mainMux.HandleFunc("POST /api/create-pdf", createHandler.CreatePDF)
	mainMux.HandleFunc("GET /api/requests", handlers.AuthMiddleware(requestsHandler.GetRequests))
	mainMux.HandleFunc("GET /api/pdf/{issueNumber}", handlers.AuthMiddleware(requestsHandler.DownloadPDF))
	mainMux.HandleFunc("GET /api/requests/{issueNumber}/detail", handlers.AuthMiddleware(registryHandler.GetDetail))
	mainMux.HandleFunc("PATCH /api/requests/{issueNumber}/detail", handlers.AuthMiddleware(registryHandler.UpdateDetail))
	mainMux.HandleFunc("PATCH /api/requests/{issueNumber}/approve", handlers.AuthMiddleware(requestsHandler.Approve))
	mainMux.HandleFunc("PATCH /api/requests/{issueNumber}/reject", handlers.AuthMiddleware(requestsHandler.Reject))
	mainMux.HandleFunc("GET /api/registry/preview", handlers.AuthMiddleware(registryHandler.PreviewRegistry))
	mainMux.HandleFunc("GET /api/registry", handlers.AuthMiddleware(registryHandler.DownloadRegistry))
	mainMux.HandleFunc("GET /api/sessions", handlers.AuthMiddleware(sessionsHandler.GetSessions))
	mainMux.HandleFunc("POST /api/sessions", handlers.AuthMiddleware(sessionsHandler.SaveSessions))
	mainMux.HandleFunc("GET /api/status/{studentCard}", statusHandler.GetStatus)

	log.Print("Server started on port :8080")
	if err := http.ListenAndServe(":8080", mainMux); err != nil {
		log.Fatal("Server starting error")
	}
}
