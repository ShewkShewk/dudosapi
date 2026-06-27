CREATE OR REPLACE FUNCTION delete_missing_students()
    RETURNS TRIGGER AS
$$
BEGIN
    DELETE FROM students WHERE students.id NOT IN (SELECT student_id FROM student_entries);
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_deletion_of_missing_students
    AFTER DELETE
    ON student_entries
    FOR EACH STATEMENT
EXECUTE FUNCTION delete_missing_students();