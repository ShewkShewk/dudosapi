CREATE TABLE students
(
    id          SERIAL PRIMARY KEY,
    school_id   SERIAL REFERENCES schools (id) ON DELETE CASCADE,
    first_name  TEXT NOT NULL,
    middle_name TEXT NOT NULL,
    last_name   TEXT NOT NULL,
    grad_year   INT  NOT NULL
);