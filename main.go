package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/Throne-of-Doom/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const filepathRoot = "."

var apiCFG = &apiConfig{}

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
	secret := os.Getenv("SECRET")
	apiCFG.dbQueries = dbQueries
	apiCFG.PLATFORM = platform
	apiCFG.SECRET = secret
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
	mux.HandleFunc("POST /api/refresh", apiCFG.refreshHandler)
	mux.HandleFunc("POST /api/revoke", apiCFG.revokeHandler)
	mux.HandleFunc("PUT /api/users", apiCFG.UpdateUserHandler)
	mux.HandleFunc("DELETE /api/chirps/{chirpID}", apiCFG.deleteChirpsHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
