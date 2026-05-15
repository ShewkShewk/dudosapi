CREATE TABLE sections
(
    id       SERIAL PRIMARY KEY,
    round_id SERIAL REFERENCES rounds (id),
    room_id  INTEGER, -- unfortunately, sections can be associated with a room that has since been deleted.
    flight   INT NOT NULL
);