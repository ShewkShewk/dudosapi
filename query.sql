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

-- name: InsertSchool :batchexec
INSERT INTO schools(id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name;

-- name: InsertEvent :batchexec
INSERT INTO events(id, tournament_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
    SET tournament_id = EXCLUDED.tournament_id,
        name          = EXCLUDED.name;

-- name: InsertEntry :batchexec
INSERT INTO entries(id, tournament_id, event_id, code, active)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(id) DO UPDATE
    SET tournament_id = EXCLUDED.tournament_id,
        event_id      = EXCLUDED.event_id,
        code          = EXCLUDED.code,
        active        = EXCLUDED.active;

-- name: InsertStudent :batchexec
INSERT INTO students(id, school_id, first_name, middle_name, last_name, grad_year)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(id) DO UPDATE
    SET school_id   = EXCLUDED.school_id,
        first_name  = EXCLUDED.first_name,
        middle_name = EXCLUDED.middle_name,
        last_name   = EXCLUDED.last_name,
        grad_year   = EXCLUDED.grad_year;

-- name: InsertStudentEntries :batchexec
INSERT INTO student_entries(student_id, entry_id)
VALUES ($1, $2)
ON CONFLICT(student_id, entry_id) DO NOTHING;

-- name: InsertRound :batchexec
INSERT INTO rounds(id, event_id, number, start_time, published)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT(id) DO UPDATE
    SET event_id   = EXCLUDED.event_id,
        number     = EXCLUDED.number,
        start_time = EXCLUDED.start_time,
        published  = EXCLUDED.published;

-- name: InsertSites :batchexec
INSERT INTO sites(id, name)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE
    SET name = EXCLUDED.name;

-- name: InsertRooms :batchexec
INSERT INTO rooms(id, site_id, name)
VALUES ($1, $2, $3)
ON CONFLICT (id) DO UPDATE
    SET site_id = EXCLUDED.site_id,
        name    = EXCLUDED.name;

-- name: InsertSections :batchexec
INSERT INTO sections(id, round_id, room_id, flight)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE
    SET round_id = EXCLUDED.round_id,
        room_id  = EXCLUDED.room_id,
        flight   = EXCLUDED.flight;

-- name: InsertBallots :batchexec
INSERT INTO ballots(id, section_id, judge_id, side, entry_id, started, result)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT(id) DO UPDATE
    SET section_id = EXCLUDED.section_id,
        judge_id   = EXCLUDED.judge_id,
        side       = EXCLUDED.side,
        entry_id   = EXCLUDED.entry_id,
        started    = EXCLUDED.started,
        result     = EXCLUDED.result;

-- name: InsertJudges :batchexec
INSERT INTO judges(id, tournament_id, person_id, first_name, last_name, email)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT(id) DO UPDATE
    SET tournament_id = EXCLUDED.tournament_id,
        person_id     = EXCLUDED.person_id,
        first_name    = EXCLUDED.first_name,
        last_name     = EXCLUDED.last_name,
        email         = EXCLUDED.email;
