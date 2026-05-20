CREATE TABLE judges
(
    id            INTEGER PRIMARY KEY,
    tournament_id INTEGER REFERENCES tournaments (id) ON DELETE CASCADE NOT NULL,
    person_id     INTEGER                                               NOT NULL,
    first_name    TEXT                                                  NOT NULL,
    last_name     TEXT                                                  NOT NULL,
    email         TEXT
);