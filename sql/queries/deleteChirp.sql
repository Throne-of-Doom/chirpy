-- name: DeleteChirp :exec
DELETE FROM chirps
WHERE ID = $1;