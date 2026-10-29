CREATE TABLE shift_handovers (
    id text PRIMARY KEY,
    shift_id text NOT NULL REFERENCES shift_instances(id),
    created_at timestamptz NOT NULL,
    created_by text NOT NULL CHECK (length(trim(created_by)) > 0),
    snapshot jsonb NOT NULL CHECK (jsonb_typeof(snapshot) = 'object'),
    snapshot_hash text NOT NULL CHECK (snapshot_hash ~ '^[a-f0-9]{64}$')
);

CREATE INDEX shift_handovers_history
    ON shift_handovers(shift_id, created_at DESC, id DESC);

CREATE FUNCTION protect_shift_handover() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'Shift handover snapshots are immutable';
END;
$$;

CREATE TRIGGER shift_handover_immutable
    BEFORE UPDATE OR DELETE ON shift_handovers
    FOR EACH ROW EXECUTE FUNCTION protect_shift_handover();
