CREATE TABLE tournaments(
    id SERIAL PRIMARY KEY,
    raw jsonb,
    date TEXT GENERATED ALWAYS AS ( raw->>'start') STORED,
    name TEXT GENERATED ALWAYS AS ( raw->>'name' ) STORED,
    updated_time TEXT GENERATED ALWAYS AS ( raw->>'backup_created' ) STORED
);

CREATE TABLE schools(
    id SERIAL PRIMARY KEY,
    name TEXT
);