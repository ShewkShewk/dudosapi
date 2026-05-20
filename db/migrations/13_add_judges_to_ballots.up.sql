ALTER TABLE ballots
    ADD COLUMN judge_id INTEGER REFERENCES judges (id);