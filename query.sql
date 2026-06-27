-- name: GetLoadedTournaments :many
SELECT id, updated_time
FROM tournaments;

-- name: GetTournament :one
SELECT id, date, name, updated_time
FROM tournaments
WHERE id = $1;

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

-- name: GetLatestPublishedRoundsPerEvent :many
SELECT DISTINCT ON (events.name) rounds.event_id   AS event_id,
                                 events.name       AS event_name,
                                 rounds.number     AS round_number,
                                 rounds.id         AS round_id,
                                 rounds.start_time AS start_time
FROM rounds
         JOIN events
              ON events.id = rounds.event_id
WHERE events.tournament_id = $1
  AND published = TRUE
ORDER BY events.name, rounds.number DESC;

-- name: GetPairingsWithBallots :many
WITH unique_section_judges AS (SELECT DISTINCT ON (ballots.section_id, judges.id) ballots.section_id,
                                                                                  judges.id,
                                                                                  judges.first_name,
                                                                                  judges.last_name,
                                                                                  judges.person_id,
                                                                                  ballots.started
                               FROM ballots
                                        LEFT JOIN judges ON ballots.judge_id = judges.id
                                        JOIN sections ON ballots.section_id = sections.id
                               WHERE sections.round_id = ANY ($1::int[])
                                 AND judges.id IS NOT NULL),
     judges_aggregated AS (SELECT section_id,
                                  JSON_AGG(
                                          JSON_BUILD_OBJECT('id', id, 'firstName', first_name, 'lastName', last_name,
                                                            'personId', person_id, 'started', started)
                                  ) AS aggregated_judges
                           FROM unique_section_judges
                           GROUP BY section_id),
     matchups_with_ballots AS (SELECT section_id,
                                      MAX(sections.flight)                                         AS flight,
                                      MAX(rounds.event_id)                                         AS event_id,
                                      MAX(room_id)                                                 AS room_id,
                                      MAX(entry_id) FILTER (WHERE side = 'AFF')                    AS aff_team,
                                      MAX(entry_id) FILTER (WHERE side = 'NEG')                    AS neg_team,
                                      JSON_AGG(
                                      JSON_BUILD_OBJECT('side', side, 'result', result, 'judge', ballots.judge_id)
                                              ) FILTER (WHERE result IS NOT NULL AND side = 'AFF') AS aff_ballots,
                                      JSON_AGG(
                                      JSON_BUILD_OBJECT('side', side, 'result', result, 'judge', ballots.judge_id)
                                              ) FILTER (WHERE result IS NOT NULL AND side = 'NEG') AS neg_ballots
                               FROM ballots
                                        JOIN sections ON ballots.section_id = sections.id
                                        JOIN rounds ON sections.round_id = rounds.id
                               WHERE sections.round_id = ANY ($1::int[])
                               GROUP BY section_id)
SELECT matchups_with_ballots.event_id::int AS event_id,
       matchups_with_ballots.section_id    AS section_id,
       matchups_with_ballots.flight::int   AS flight,
       rooms.id                            AS room_id,
       rooms.name                          AS room_name,
       aff_entries.id                      AS aff_team_entry_id,
       aff_entries.code                    AS aff_team_entry_code,
       neg_entries.id                      AS neg_team_entry_id,
       neg_entries.code                    As neg_team_entry_code,
       matchups_with_ballots.aff_ballots   AS aff_team_ballots,
       matchups_with_ballots.neg_ballots   AS neg_team_ballots,
       judges_aggregated.aggregated_judges AS associated_judges
FROM matchups_with_ballots
         LEFT JOIN entries AS aff_entries ON aff_team = aff_entries.id
         LEFT JOIN entries AS neg_entries ON neg_team = neg_entries.id
         LEFT JOIN rooms ON room_id = rooms.id
         LEFT JOIN judges_aggregated ON matchups_with_ballots.section_id = judges_aggregated.section_id
ORDER BY room_name;