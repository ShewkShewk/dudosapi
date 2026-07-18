CREATE TABLE school_entries
(
    tournament_id SERIAL REFERENCES tournaments (id),
    school_id     SERIAL REFERENCES schools (id),
    on_site       BOOLEAN,
    PRIMARY KEY (tournament_id, school_id)
);