
CREATE TABLE departments(id text PRIMARY KEY,name text NOT NULL);
CREATE TABLE employees(id text PRIMARY KEY,department_id text NOT NULL REFERENCES departments,role text NOT NULL);
CREATE TABLE production_lines(id text PRIMARY KEY,department_id text NOT NULL REFERENCES departments,name text NOT NULL,facility_id text NOT NULL DEFAULT 'FACTORY-01');
CREATE TABLE machines(id text PRIMARY KEY,line_id text NOT NULL REFERENCES production_lines,name text NOT NULL);
CREATE TABLE calendar_days(day date PRIMARY KEY,is_working boolean NOT NULL,description text NOT NULL);
CREATE TABLE shift_instances(id text PRIMARY KEY,line_id text NOT NULL REFERENCES production_lines,operational_date date NOT NULL REFERENCES calendar_days(day),code text NOT NULL CHECK(code IN('S1','S2','S3')),starts_at timestamptz NOT NULL,ends_at timestamptz NOT NULL,required_staff integer NOT NULL CHECK(required_staff>=0),product text NOT NULL,complete boolean NOT NULL DEFAULT true,CHECK(ends_at>starts_at),UNIQUE(line_id,operational_date,code));
CREATE INDEX shift_line_time ON shift_instances(line_id,starts_at);
CREATE TABLE shift_assignments(id text PRIMARY KEY,shift_id text NOT NULL REFERENCES shift_instances,employee_id text NOT NULL REFERENCES employees,UNIQUE(shift_id,employee_id));
CREATE INDEX assignment_employee ON shift_assignments(employee_id,shift_id);
CREATE TABLE attendance_records(id text PRIMARY KEY,assignment_id text NOT NULL UNIQUE REFERENCES shift_assignments,check_in timestamptz,check_out timestamptz,verified boolean NOT NULL,source_complete boolean NOT NULL,CHECK(check_in IS NULL OR check_out IS NULL OR check_out>check_in));
CREATE TABLE leave_records(id text PRIMARY KEY,employee_id text NOT NULL REFERENCES employees,starts_at timestamptz NOT NULL,ends_at timestamptz NOT NULL,approved boolean NOT NULL,CHECK(ends_at>starts_at));
CREATE INDEX leave_employee_time ON leave_records(employee_id,starts_at,ends_at);
CREATE TABLE overtime_records(id text PRIMARY KEY,assignment_id text NOT NULL REFERENCES shift_assignments,minutes integer NOT NULL CHECK(minutes>=0),cost_rupiah bigint NOT NULL CHECK(cost_rupiah>=0),approved boolean NOT NULL,complete boolean NOT NULL);
CREATE TABLE production_runs(id text PRIMARY KEY,shift_id text NOT NULL UNIQUE REFERENCES shift_instances,planned integer NOT NULL CHECK(planned>=0),actual integer NOT NULL CHECK(actual>=0),complete boolean NOT NULL);
CREATE TABLE downtime_events(id text PRIMARY KEY,shift_id text NOT NULL REFERENCES shift_instances,machine_id text NOT NULL REFERENCES machines,starts_at timestamptz NOT NULL,ends_at timestamptz NOT NULL,reason text NOT NULL,CHECK(ends_at>starts_at));
CREATE TABLE expenses(id text PRIMARY KEY,line_id text NOT NULL REFERENCES production_lines,shift_id text NOT NULL REFERENCES shift_instances,category text NOT NULL,amount_rupiah bigint NOT NULL CHECK(amount_rupiah>=0),occurred_at timestamptz NOT NULL,complete boolean NOT NULL);
CREATE INDEX expense_line_time ON expenses(line_id,occurred_at);
CREATE TABLE budgets(id text PRIMARY KEY,line_id text NOT NULL REFERENCES production_lines,category text NOT NULL,period_start date NOT NULL,period_end date NOT NULL,amount_rupiah bigint NOT NULL CHECK(amount_rupiah>=0),CHECK(period_end>=period_start));
