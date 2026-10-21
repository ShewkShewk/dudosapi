CREATE TABLE entry_awards
(
    tournament_id SERIAL REFERENCES tournaments (id) ON DELETE CASCADE,
    event_id      SERIAL REFERENCES events (id) ON DELETE CASCADE,
    rank          INT NOT NULL,
    entry_id      SERIAL,
    PRIMARY KEY (tournament_id, event_id, rank)
);