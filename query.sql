-- name: GetLoadedTournaments :many
SELECT id, name FROM tournaments;

-- name: LoadTournament :exec
INSERT INTO tournaments (id, raw)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET raw = EXCLUDED.raw;