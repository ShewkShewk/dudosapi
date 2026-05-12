CREATE TABLE student_entries
(
    student_id SERIAL REFERENCES students (id) ON DELETE CASCADE,
    entry_id   SERIAL REFERENCES entries (id) ON DELETE CASCADE,
    PRIMARY KEY (student_id, entry_id)
);