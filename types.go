package main

import (
"sync/atomic"
"time"

"github.com/Throne-of-Doom/chirpy/internal/database"
"github.com/google/uuid"
)

type apiConfig struct {
fileserverHits atomic.Int32
dbQueries      *database.Queries
PLATFORM       string
}

type ChirpResponse struct {
ID        uuid.UUID `json:"id"`
CreatedAt time.Time `json:"created_at"`
UpdatedAt time.Time `json:"updated_at"`
Body      string    `json:"body"`
UserID    uuid.UUID `json:"user_id"`
}
