package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Throne-of-Doom/chirpy/internal/auth"
	"github.com/Throne-of-Doom/chirpy/internal/database"
)

func (cfg *apiConfig) loginHandler(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var params CredentialsRequest
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	dbUser, err := cfg.dbQueries.GetUserByEmail(r.Context(), params.Email)
	if err != nil {
		respondWithError(w, 401, "incorrect email or password", err)
		return
	}
	ok, err := auth.CheckPasswordHash(params.Password, dbUser.HashedPassword)
	if err != nil || !ok {
		respondWithError(w, 401, "incorrect email or password", err)
		return
	}

	expiresIn := time.Hour
	token, err := auth.MakeJWT(dbUser.ID, cfg.SECRET, expiresIn)
	if err != nil {
		respondWithError(w, 500, "couldn't create token", err)
		return
	}

	refreshToken, err := auth.MakeRefreshToken()
	if err != nil {
		respondWithError(w, 500, "couldn't create refresh token", err)
		return
	}

	expiresAt := time.Now().UTC().Add(60 * 24 * time.Hour)

	_, err = cfg.dbQueries.CreateRefreshToken(r.Context(), database.CreateRefreshTokenParams{
		Token:     refreshToken,
		UserID:    dbUser.ID,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		respondWithError(w, 500, "couldn't save refresh token", err)
		return
	}
	resp := loginResponse{
		ID:           dbUser.ID,
		CreatedAt:    dbUser.CreatedAt,
		UpdatedAt:    dbUser.UpdatedAt,
		Email:        dbUser.Email,
		Token:        token,
		RefreshToken: refreshToken,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) refreshHandler(w http.ResponseWriter, r *http.Request) {
	type refreshResponse struct {
		Token string `json:"token"`
	}
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authorization header", err)
		return
	}

	user, err := cfg.dbQueries.GetUserFromRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "invalid or expired refresh token", err)
		return
	}

	newToken, err := auth.MakeJWT(user.ID, cfg.SECRET, time.Hour)
	if err != nil {
		respondWithError(w, 500, "couldn't create token", err)
		return
	}

	resp := refreshResponse{
		Token: newToken,
	}
	respondWithJSON(w, 200, resp)
}

func (cfg *apiConfig) revokeHandler(w http.ResponseWriter, r *http.Request) {
	refreshToken, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authorization header", err)
		return
	}

	err = cfg.dbQueries.RevokeRefreshToken(r.Context(), refreshToken)
	if err != nil {
		respondWithError(w, 401, "invalid refresh token", err)
		return
	}

	w.WriteHeader(204)
}
