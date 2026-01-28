package main

import (
	"database/sql"
	"encoding/json"
	"github.com/Throne-of-Doom/chirpy/internal/auth"
	"github.com/google/uuid"
	"net/http"
)

type Data struct {
	UserID uuid.UUID `json:"user_id"`
}

type eventPolka struct {
	Event string `json:"event"`
	Data  Data   `json:"data"`
}

func (cfg *apiConfig) upgradeChirpyHandler(w http.ResponseWriter, r *http.Request) {
	APIKey, err := auth.GetAPIKey(r.Header)
	if err != nil {
		respondWithError(w, 401, "invalid authorization header", err)
		return
	}
	if APIKey != cfg.POLKA_KEY {
		respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
		return

	}
	decoder := json.NewDecoder(r.Body)
	var params eventPolka
	err = decoder.Decode(&params)
	if err != nil {
		respondWithError(w, 400, "unable to decode JSON", err)
		return
	}
	if params.Event != "user.upgraded" {
		w.WriteHeader(204)
		return
	}
	err = cfg.dbQueries.UpgradeUser(r.Context(), params.Data.UserID)
	if err == sql.ErrNoRows {
		respondWithError(w, 404, "unable to upgrade user", err)
		return
	}
	if err != nil {
		respondWithError(w, 500, "error updating user", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
