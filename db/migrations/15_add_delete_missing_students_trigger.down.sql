DROP TRIGGER IF EXISTS trigger_deletion_of_missing_students ON student_entries;

DROP FUNCTION IF EXISTS delete_missing_students();