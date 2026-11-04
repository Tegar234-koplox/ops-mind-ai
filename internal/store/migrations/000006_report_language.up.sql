ALTER TABLE investigation_runs
    ADD COLUMN report_language text NOT NULL DEFAULT 'id'
    CHECK (report_language IN ('id', 'en'));
