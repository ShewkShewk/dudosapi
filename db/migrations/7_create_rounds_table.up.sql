CREATE TABLE rounds
(
    id         SERIAL PRIMARY KEY,
    event_id   SERIAL REFERENCES events (id) ON DELETE CASCADE,
    number     INT       NOT NULL,
    start_time TIMESTAMP NOT NULL,
    published  BOOL      NOT NULL
);