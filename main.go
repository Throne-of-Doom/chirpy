package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Throne-of-Doom/chirpy/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const filepathRoot = "."

var apiCFG = &apiConfig{}

type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type user struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Email     string    `json:"email"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("An error has occured %s", err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("An error has occured", err)
	}
	defer db.Close()
	dbQueries := database.New(db)

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	apiCFG.dbQueries = dbQueries
	apiCFG.PLATFORM = platform
	mux := http.NewServeMux()
	mux.Handle("/app/", apiCFG.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))
	mux.HandleFunc("GET /api/healthz", readinessHandler)
	mux.HandleFunc("GET /admin/metrics", apiCFG.metricsHandler)
	mux.HandleFunc("GET /api/chirps", apiCFG.getChirpsHandler)
	mux.HandleFunc("POST /admin/reset", apiCFG.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiCFG.createChirpsHandler)
	mux.HandleFunc("POST /api/users", apiCFG.createUserHandler)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCFG.getChirpHandler)
	mux.HandleFunc("POST /api/login", apiCFG.loginHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
