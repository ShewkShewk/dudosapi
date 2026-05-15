CREATE TYPE ballot_side AS ENUM ('AFF', 'NEG');

CREATE TYPE ballot_result AS ENUM ('WIN', 'LOSS', 'BYE', 'FFT');

CREATE TABLE ballots
(
    id         INTEGER PRIMARY KEY,
    section_id INTEGER REFERENCES sections (id),
    side       ballot_side,
    entry_id   INTEGER REFERENCES entries (id),
    result     ballot_result
);