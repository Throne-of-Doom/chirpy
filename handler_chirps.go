package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/Throne-of-Doom/chirpy/internal/auth"
	"github.com/Throne-of-Doom/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) createChirpsHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid or missing token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRET)
	if err != nil {
		respondWithError(w, 401, "invalid or expired token", err)
		return
	}
	type data struct {
		Body string `json:"body"`
	}
	profaneWords := map[string]struct{}{
		"kerfuffle": {},
		"sharbert":  {},
		"fornax":    {},
	}
	decoder := json.NewDecoder(r.Body)
	params := data{}
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	if len(params.Body) > 140 {
		err := errors.New("chirp is too long, limit 140 characters")
		respondWithError(w, 400, "Chirp is too long, Limit 140 Characters", err)
		return
	}

	cleaned := profaneReplace(params.Body, profaneWords)

	dbChirp, err := cfg.dbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
		Body:   cleaned,
		UserID: userID,
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

func (cfg *apiConfig) getChirpsHandler(w http.ResponseWriter, r *http.Request) {
	chirps, err := cfg.dbQueries.GetChirps(r.Context())
	if err != nil {
		respondWithError(w, 500, "an error has occured", err)
		return
	}
	urlSort := ""
	authorIDStr := r.URL.Query().Get("author_id")
	urlSort = r.URL.Query().Get("sort")
	var authorUUID uuid.UUID
	var filterByAuthor bool = false

	if authorIDStr != "" {
		authorUUID, err = uuid.Parse(authorIDStr)

		if err != nil {
			respondWithError(w, 400, "Invalid author_id", err)
			return
		}
		filterByAuthor = true
	}

	responseChirps := []ChirpResponse{}

	for _, chirp := range chirps {
		if filterByAuthor && chirp.UserID != authorUUID {
			continue
		}
		responseChirps = append(responseChirps, ChirpResponse{
			ID:        chirp.ID,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
			Body:      chirp.Body,
			UserID:    chirp.UserID,
		})
	}
	if urlSort == "asc" || urlSort == "" {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.Before(responseChirps[j].CreatedAt)
		})
	} else if urlSort == "desc" {
		sort.Slice(responseChirps, func(i, j int) bool {
			return responseChirps[i].CreatedAt.After(responseChirps[j].CreatedAt)
		})
	}
	respondWithJSON(w, 200, responseChirps)
}

func (cfg *apiConfig) getChirpHandler(w http.ResponseWriter, r *http.Request) {
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, 400, "invalid chirp ID", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found", err)
		return
	}
	response := ChirpResponse{
		ID:        chirp.ID,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
		Body:      chirp.Body,
		UserID:    chirp.UserID,
	}

	respondWithJSON(w, 200, response)
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

func (cfg *apiConfig) deleteChirpsHandler(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid or missing token", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.SECRET)
	if err != nil {
		respondWithError(w, 401, "invalid or expired token", err)
		return
	}
	chirpIDStr := r.PathValue("chirpID")
	chirpID, err := uuid.Parse(chirpIDStr)
	if err != nil {
		respondWithError(w, 400, "invalid chirp ID", err)
		return
	}
	chirp, err := cfg.dbQueries.GetChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "chirp not found", err)
		return
	}
	if userID != chirp.UserID {
		respondWithError(w, 403, "cannot delete chirp", nil)
		return
	}
	err = cfg.dbQueries.DeleteChirp(r.Context(), chirpID)
	if err != nil {
		respondWithError(w, 404, "cannot delete chirp", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
