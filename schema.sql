CREATE TABLE tournaments
(
    id           SERIAL PRIMARY KEY,
    raw          jsonb,
    date         TEXT GENERATED ALWAYS AS ( raw ->> 'start') STORED,
    name         TEXT GENERATED ALWAYS AS ( raw ->> 'name' ) STORED,
    updated_time TEXT GENERATED ALWAYS AS ( raw ->> 'backup_created' ) STORED
);

CREATE TABLE schools
(
    id   SERIAL PRIMARY KEY,
    name TEXT
);

CREATE TABLE events
(
    id            SERIAL PRIMARY KEY,
    tournament_id SERIAL REFERENCES tournaments (id) ON DELETE CASCADE,
    name          TEXT
);

CREATE TABLE entries
(
    id            SERIAL,
    tournament_id SERIAL REFERENCES tournaments (id) ON DELETE CASCADE,
    event_id      SERIAL REFERENCES events (id) ON DELETE CASCADE,
    code          TEXT,
    active        BOOLEAN
);

CREATE TABLE students
(
    id          SERIAL PRIMARY KEY,
    school_id   SERIAL REFERENCES schools (id) ON DELETE CASCADE,
    first_name  TEXT NOT NULL,
    middle_name TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    grad_year   INT  NOT NULL
);