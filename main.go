package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	"github.com/Throne-of-Doom/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const filepathRoot = "."

var apiCFG = &apiConfig{}

type apiConfig struct {
	fileserverHits atomic.Int32
	dbQueries      *database.Queries
	PLATFORM       string
}

func readinessHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.fileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) metricsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	hits := cfg.fileserverHits.Load()
	html := fmt.Sprintf(
		`<html>
  			<body>
    			<h1>Welcome, Chirpy Admin</h1>
    			<p>Chirpy has been visited %d times!</p>
  			</body>
		</html>`, hits,
	)
	w.Write([]byte(html))
}

func (cfg *apiConfig) resetHandler(w http.ResponseWriter, r *http.Request) {
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	cfg.fileserverHits.Store(0)
	err := cfg.dbQueries.DeleteUsers(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "couldn't delete users", err)
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func respondWithError(w http.ResponseWriter, code int, msg string, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(struct {
		Error string `json:"error"`
	}{Error: msg})
	log.Printf("An error has occured: %s, %s", err, msg)

}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	dat, err := json.Marshal(payload)
	if err != nil {
		log.Printf("Error marshalling JSON: %s", err)
		w.WriteHeader(500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(dat)
}

func profaneReplace(msg string, profaneWords map[string]struct{}) string {
	censor := "****"
	words := strings.Split(msg, " ")
	for i, word := range words {
		lowered := strings.ToLower(word)
		if _, ok := profaneWords[lowered]; ok {
			words[i] = censor
		}
	}
	cleaned := strings.Join(words, " ")
	return cleaned
}

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {
	type email struct {
		Email string `json:"email"`
	}
	decoder := json.NewDecoder(r.Body)
	params := email{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 500, "couldn't call database", err)
		return
	}

	type user struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Email     string    `json:"email"`
	}
	resp := user{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, 201, resp)

}

func (cfg *apiConfig) createChirpsHandler(w http.ResponseWriter, r *http.Request) {
	type data struct {
		Body   string    `json:"body"`
		UserID uuid.UUID `json:"user_id"`
	}
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	decoder := json.NewDecoder(r.Body)
	params := data{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	if len(params.Body) > 140 {
		respondWithError(w, 400, "Chirp is too long, Limit 140 Characters", nil)
		return
	}

	cleaned := profaneReplace(params.Body, profaneWords)

	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: params.UserID,
	})
	if err != nil {
		respondWithError(w, 500, "couldn't call database", err)
		return
	}

	type chirp struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
		Body      string    `json:"body"`
		UserID    uuid.UUID `json:"user_id"`
	}

	resp := chirp{
		ID:        dbChirp.ID,
		CreatedAt: dbChirp.CreatedAt,
		UpdatedAt: dbChirp.UpdatedAt,
		Body:      dbChirp.Body,
		UserID:    dbChirp.UserID,
	}
	respondWithJSON(w, http.StatusCreated, resp)
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
	mux.HandleFunc("POST /admin/reset", apiCFG.resetHandler)
	mux.HandleFunc("POST /api/chirps", apiCFG.createChirpsHandler)
	mux.HandleFunc("POST /api/users", apiCFG.createUserHandler)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	srv.ListenAndServe()
}
