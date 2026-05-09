-- name: GetLoadedTournaments :many
SELECT id, updated_time
FROM tournaments;

-- name: LoadTournament :exec
INSERT INTO tournaments (id, raw)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET raw = EXCLUDED.raw;

-- name: DeleteTournament :exec
DELETE
FROM tournaments
WHERE id = $1;

-- name: InsertSchool :exec
INSERT INTO schools(id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name;

-- name: InsertEvent :exec
INSERT INTO events(id, tournament_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
    SET tournament_id = EXCLUDED.tournament_id,
        name          = EXCLUDED.name;