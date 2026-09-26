ALTER TABLE school_entries
    DROP CONSTRAINT school_entries_tournament_id_fkey;
ALTER TABLE school_entries
    ADD CONSTRAINT school_entries_tournament_id_fkey FOREIGN KEY (tournament_id) REFERENCES tournaments (id);
