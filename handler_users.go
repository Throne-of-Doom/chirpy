package main

import (
	"encoding/json"
	"net/http"

	"github.com/Throne-of-Doom/chirpy/internal/auth"
	"github.com/Throne-of-Doom/chirpy/internal/database"
	"github.com/google/uuid"
)

func (cfg *apiConfig) createUserHandler(w http.ResponseWriter, r *http.Request) {

	decoder := json.NewDecoder(r.Body)
	params := CredentialsRequest{}
	err := decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	hashedPW, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "an error has ocurred", err)
		return
	}

	dbUser, err := cfg.dbQueries.CreateUser(r.Context(), database.CreateUserParams{
		Email:          params.Email,
		HashedPassword: hashedPW,
	})
	if err != nil {
		respondWithError(w, 500, "couldn't call database", err)
		return
	}

	resp := user{
		ID:        dbUser.ID,
		CreatedAt: dbUser.CreatedAt,
		UpdatedAt: dbUser.UpdatedAt,
		Email:     dbUser.Email,
	}
	respondWithJSON(w, 201, resp)
}

func (cfg *apiConfig) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	type updateUser struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid or missing token", err)
		return
	}
	decoder := json.NewDecoder(r.Body)
	params := &updateUser{}
	err = decoder.Decode(params)
	if err != nil {
		respondWithError(w, 500, "couldn't decode parameters", err)
		return
	}
	hashed_password, err := auth.HashPassword(params.Password)
	if err != nil {
		respondWithError(w, 500, "couldn't hash password", err)
		return
	}

	userID, err := auth.ValidateJWT(token, apiCFG.SECRET)
	if err != nil {
		respondWithError(w, 401, "Not authorized", err)
		return
	}
	user_update := database.UpdateUserParams{
		ID:             userID,
		Email:          params.Email,
		HashedPassword: hashed_password,
	}
	update, err := cfg.dbQueries.UpdateUser(r.Context(), user_update)
	if err != nil {
		respondWithError(w, 500, "error updating user", err)
		return
	}
	type response struct {
		ID    uuid.UUID `json:"user_id"`
		Email string    `json:"email"`
	}

	resp := response{
		ID:    update.ID,
		Email: update.Email,
	}
	respondWithJSON(w, 200, resp)
}
