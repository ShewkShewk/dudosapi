CREATE TABLE sections
(
    id       SERIAL PRIMARY KEY,
    round_id SERIAL REFERENCES rounds (id),
    room_id  INTEGER REFERENCES rooms (id),
    flight   INT NOT NULL
);