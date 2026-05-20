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

CREATE TABLE student_entries
(
    student_id SERIAL REFERENCES students (id) ON DELETE CASCADE,
    entry_id   SERIAL REFERENCES entries (id) ON DELETE CASCADE,
    PRIMARY KEY (student_id, entry_id)
);

CREATE TABLE rounds
(
    id         SERIAL PRIMARY KEY,
    event_id   SERIAL REFERENCES events (id) ON DELETE CASCADE,
    number     INT       NOT NULL,
    start_time TIMESTAMP NOT NULL,
    published  BOOL      NOT NULL
);

CREATE TABLE sites
(
    id   SERIAL PRIMARY KEY,
    NAME TEXT
);

CREATE TABLE rooms
(
    id      SERIAL PRIMARY KEY,
    site_id SERIAL REFERENCES sites (id) ON DELETE CASCADE,
    name    TEXT
);

CREATE TABLE sections
(
    id       SERIAL PRIMARY KEY,
    round_id SERIAL REFERENCES rounds (id),
    room_id  INTEGER REFERENCES rooms (id),
    flight   INT NOT NULL
);

CREATE TYPE ballot_side AS ENUM ('AFF', 'NEG');

CREATE TYPE ballot_result AS ENUM ('WIN', 'LOSS', 'BYE', 'FFT');

CREATE TABLE ballots
(
    id         INTEGER PRIMARY KEY,
    section_id INTEGER REFERENCES sections (id),
    side       ballot_side,
    entry_id   INTEGER REFERENCES entries (id),
    started    BOOLEAN,
    result     ballot_result
);

CREATE TABLE judges
(
    id            INTEGER PRIMARY KEY,
    tournament_id INTEGER REFERENCES tournaments (id) ON DELETE CASCADE NOT NULL,
    person_id     INTEGER                                               NOT NULL,
    first_name    TEXT                                                  NOT NULL,
    last_name     TEXT                                                  NOT NULL,
    email         TEXT
);