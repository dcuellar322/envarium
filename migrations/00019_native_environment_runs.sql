-- +goose Up
ALTER TABLE runs ADD COLUMN runtime_scope_id TEXT NOT NULL DEFAULT '';
UPDATE runs SET runtime_scope_id = project_id;
CREATE INDEX runs_runtime_scope_idx ON runs(runtime_scope_id, (ended_at IS NULL) DESC, started_at DESC, id DESC);
ALTER TABLE health_samples ADD COLUMN runtime_scope_id TEXT NOT NULL DEFAULT '';
UPDATE health_samples SET runtime_scope_id = project_id;
CREATE INDEX health_runtime_scope_idx ON health_samples(runtime_scope_id, service_id, check_id, observed_at DESC, id DESC);
UPDATE system_health SET schema_version = 19 WHERE singleton = 1;

-- +goose Down
-- Refuse a downgrade that would conflate isolated run ownership.
CREATE TABLE native_scope_downgrade_guard (unsupported INTEGER CHECK(unsupported = 0));
INSERT INTO native_scope_downgrade_guard SELECT count(*) FROM runs WHERE runtime_scope_id != project_id;
INSERT INTO native_scope_downgrade_guard SELECT count(*) FROM health_samples WHERE runtime_scope_id != project_id;
DROP TABLE native_scope_downgrade_guard;
DROP INDEX health_runtime_scope_idx;
ALTER TABLE health_samples DROP COLUMN runtime_scope_id;
DROP INDEX runs_runtime_scope_idx;
ALTER TABLE runs DROP COLUMN runtime_scope_id;
UPDATE system_health SET schema_version = 18 WHERE singleton = 1;
