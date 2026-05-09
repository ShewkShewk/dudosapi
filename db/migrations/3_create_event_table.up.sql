CREATE TABLE events
(
    id            SERIAL PRIMARY KEY,
    tournament_id SERIAL REFERENCES tournaments (id) ON DELETE CASCADE,
    name          TEXT
);