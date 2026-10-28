ALTER TABLE attendance_records
    ADD COLUMN absence_confirmed_at timestamptz,
    ADD COLUMN absence_confirmed_by text;

ALTER TABLE attendance_records ADD CONSTRAINT confirmed_absence_source CHECK (
    (absence_confirmed_at IS NULL AND absence_confirmed_by IS NULL)
    OR (
        absence_confirmed_at IS NOT NULL
        AND absence_confirmed_by IS NOT NULL
        AND length(trim(absence_confirmed_by)) > 0
        AND check_in IS NULL AND check_out IS NULL
        AND NOT verified AND source_complete
    )
);

CREATE TABLE expense_closures (
    shift_id text PRIMARY KEY REFERENCES shift_instances,
    closed_at timestamptz NOT NULL,
    closed_by text NOT NULL CHECK (length(trim(closed_by)) > 0)
);

CREATE TABLE overtime_approval_events (
    id text PRIMARY KEY,
    overtime_id text NOT NULL REFERENCES overtime_records,
    status text NOT NULL CHECK (status IN ('approved', 'revoked')),
    occurred_at timestamptz NOT NULL,
    actor text NOT NULL CHECK (length(trim(actor)) > 0),
    UNIQUE (overtime_id, occurred_at)
);
CREATE INDEX approval_record_time
    ON overtime_approval_events(overtime_id, occurred_at DESC, id DESC);

CREATE FUNCTION prevent_approval_history_change() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'overtime_approval_events append-only';
END;
$$;
CREATE TRIGGER approval_history_immutable
    BEFORE UPDATE OR DELETE ON overtime_approval_events
    FOR EACH ROW EXECUTE FUNCTION prevent_approval_history_change();
