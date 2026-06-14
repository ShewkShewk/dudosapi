ALTER TABLE sections
    DROP CONSTRAINT sections_round_id_fkey;
ALTER TABLE sections
    ADD CONSTRAINT sections_round_id_fkey FOREIGN KEY (round_id) REFERENCES rounds (id);

ALTER TABLE ballots
    DROP CONSTRAINT ballots_section_id_fkey;
ALTER TABLE ballots
    ADD CONSTRAINT ballots_section_id_fkey FOREIGN KEY (section_id) REFERENCES sections (id);

ALTER TABLE ballots
    DROP CONSTRAINT ballots_judge_id_fkey;
ALTER TABLE ballots
    ADD CONSTRAINT ballots_judge_id_fkey FOREIGN KEY (judge_id) REFERENCES judges (id);

ALTER TABLE ballots
    DROP CONSTRAINT ballots_entry_id_fkey;
ALTER TABLE ballots
    ADD CONSTRAINT ballots_entry_id_fkey FOREIGN KEY (entry_id) REFERENCES entries (id);