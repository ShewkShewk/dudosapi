CREATE TABLE entries
(
    id            SERIAL PRIMARY KEY,
    tournament_id SERIAL REFERENCES tournaments (id) ON DELETE CASCADE,
    event_id      SERIAL REFERENCES events (id) ON DELETE CASCADE,
    code          TEXT,
    active        BOOLEAN
);