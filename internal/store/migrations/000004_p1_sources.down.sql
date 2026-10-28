DROP TABLE IF EXISTS overtime_approval_events, expense_closures;
DROP FUNCTION IF EXISTS prevent_approval_history_change();
ALTER TABLE attendance_records
    DROP CONSTRAINT confirmed_absence_source,
    DROP COLUMN absence_confirmed_at,
    DROP COLUMN absence_confirmed_by;
