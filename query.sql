-- name: GetLoadedTournaments :many
SELECT id, updated_time FROM tournaments;

-- name: LoadTournament :exec
INSERT INTO tournaments (id, raw)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET raw = EXCLUDED.raw;

-- name: DeleteTournament :exec
DELETE FROM tournaments WHERE id = $1;