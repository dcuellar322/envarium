-- +goose Up
CREATE TABLE runs_v18 (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service_id TEXT NOT NULL,
    runtime_driver TEXT NOT NULL,
    origin TEXT NOT NULL CHECK (origin IN ('envarium', 'external')),
    started_at TEXT NOT NULL,
    ended_at TEXT,
    exit_code INTEGER,
    termination_reason TEXT NOT NULL DEFAULT '',
    identity_fingerprint TEXT NOT NULL,
    restart_count INTEGER NOT NULL DEFAULT 0 CHECK (restart_count >= 0),
    operation_id TEXT
);

CREATE TABLE run_processes_v18 (
    run_id TEXT NOT NULL REFERENCES runs_v18(id) ON DELETE CASCADE,
    pid INTEGER NOT NULL,
    process_group_id INTEGER NOT NULL,
    executable TEXT NOT NULL,
    started_at TEXT NOT NULL,
    working_directory TEXT NOT NULL,
    identity_fingerprint TEXT NOT NULL,
    observed_at TEXT NOT NULL,
    PRIMARY KEY (run_id, pid, started_at)
);

INSERT INTO runs_v18 (
    id, project_id, service_id, runtime_driver, origin, started_at, ended_at,
    exit_code, termination_reason, identity_fingerprint, restart_count, operation_id
)
SELECT
    id, project_id, service_id, runtime_driver,
    CASE origin WHEN 'switchyard' THEN 'envarium' ELSE origin END,
    started_at, ended_at, exit_code, termination_reason, identity_fingerprint,
    restart_count, operation_id
FROM runs;

INSERT INTO run_processes_v18 (
    run_id, pid, process_group_id, executable, started_at, working_directory,
    identity_fingerprint, observed_at
)
SELECT
    run_id, pid, process_group_id, executable, started_at,
    replace(working_directory, '/switchyard', '/envarium'),
    identity_fingerprint, observed_at
FROM run_processes;

DROP TABLE run_processes;
DROP TABLE runs;
ALTER TABLE runs_v18 RENAME TO runs;
ALTER TABLE run_processes_v18 RENAME TO run_processes;

CREATE INDEX runs_project_active_idx ON runs (project_id, ended_at, service_id);
CREATE INDEX run_processes_group_idx ON run_processes (process_group_id, run_id);

UPDATE projects
SET slug = CASE slug WHEN 'switchyard' THEN 'envarium' ELSE slug END,
    display_name = CASE display_name WHEN 'Switchyard' THEN 'Envarium' ELSE display_name END,
    description = replace(description, 'Switchyard', 'Envarium'),
    primary_location = replace(primary_location, '/switchyard', '/envarium')
WHERE slug = 'switchyard'
   OR display_name = 'Switchyard'
   OR primary_location LIKE '%/switchyard';

UPDATE project_locations
SET path = replace(path, '/switchyard', '/envarium')
WHERE path LIKE '%/switchyard';

UPDATE manifest_snapshots
SET manifest_json = replace(
    replace(
        replace(manifest_json, 'switchyard.dev', 'envarium.dev'),
        '.switchyard', '.envarium'
    ),
    'Switchyard', 'Envarium'
)
WHERE manifest_json LIKE '%switchyard%'
   OR manifest_json LIKE '%Switchyard%';

UPDATE manifest_proposals
SET schema_version = replace(schema_version, 'switchyard.dev', 'envarium.dev'),
    candidate_json = replace(
        replace(
            replace(candidate_json, 'switchyard.dev', 'envarium.dev'),
            '.switchyard', '.envarium'
        ),
        'Switchyard', 'Envarium'
    ),
    confidence_json = replace(confidence_json, 'switchyard.dev', 'envarium.dev'),
    unresolved_json = replace(unresolved_json, 'switchyard.dev', 'envarium.dev'),
    validation_json = replace(validation_json, 'switchyard.dev', 'envarium.dev')
WHERE schema_version LIKE '%switchyard%'
   OR candidate_json LIKE '%switchyard%'
   OR candidate_json LIKE '%Switchyard%'
   OR confidence_json LIKE '%switchyard%'
   OR unresolved_json LIKE '%switchyard%'
   OR validation_json LIKE '%switchyard%';

-- The repository formerly named Switchyard can contain its own executable,
-- socket, data-directory, and project identifiers in accepted manifests.
-- Restrict the broad lowercase rename to that project; other projects may
-- legitimately refer to an unrelated product with the same word.
UPDATE manifest_snapshots
SET manifest_json = replace(manifest_json, 'switchyard', 'envarium')
WHERE project_id IN (SELECT id FROM projects WHERE slug = 'envarium')
  AND manifest_json LIKE '%switchyard%';

UPDATE manifest_proposals
SET schema_version = replace(schema_version, 'switchyard', 'envarium'),
    candidate_json = replace(candidate_json, 'switchyard', 'envarium'),
    confidence_json = replace(confidence_json, 'switchyard', 'envarium'),
    unresolved_json = replace(unresolved_json, 'switchyard', 'envarium'),
    validation_json = replace(validation_json, 'switchyard', 'envarium')
WHERE project_id IN (SELECT id FROM projects WHERE slug = 'envarium')
  AND (
      schema_version LIKE '%switchyard%'
      OR candidate_json LIKE '%switchyard%'
      OR confidence_json LIKE '%switchyard%'
      OR unresolved_json LIKE '%switchyard%'
      OR validation_json LIKE '%switchyard%'
  );

UPDATE discovery_evidence
SET source_path = replace(source_path, '/switchyard', '/envarium'),
    data_json = replace(
        replace(data_json, 'switchyard.dev', 'envarium.dev'),
        '.switchyard', '.envarium'
    ),
    warnings_json = replace(warnings_json, 'switchyard.dev', 'envarium.dev')
WHERE source_path LIKE '%/switchyard%'
   OR data_json LIKE '%switchyard%'
   OR warnings_json LIKE '%switchyard%';

UPDATE project_environments
SET path = replace(path, '/switchyard', '/envarium'),
    compose_project_name = replace(compose_project_name, 'switchyard', 'envarium')
WHERE path LIKE '%/switchyard%'
   OR compose_project_name LIKE '%switchyard%';

UPDATE plugin_registrations
SET manifest_path = replace(manifest_path, '/Switchyard/', '/Envarium/'),
    executable_path = replace(executable_path, '/Switchyard/', '/Envarium/'),
    arguments_json = replace(arguments_json, '/Switchyard/', '/Envarium/')
WHERE manifest_path LIKE '%/Switchyard/%'
   OR executable_path LIKE '%/Switchyard/%'
   OR arguments_json LIKE '%/Switchyard/%';

UPDATE log_segments
SET path = replace(path, '/Switchyard/', '/Envarium/')
WHERE path LIKE '%/Switchyard/%';

UPDATE terminal_sessions
SET working_directory = replace(working_directory, '/switchyard', '/envarium')
WHERE working_directory LIKE '%/switchyard%';

UPDATE settings
SET document_json = replace(
    replace(
        replace(document_json, 'switchyard.dev', 'envarium.dev'),
        '.switchyard', '.envarium'
    ),
    'Switchyard', 'Envarium'
)
WHERE document_json LIKE '%switchyard%'
   OR document_json LIKE '%Switchyard%';

UPDATE system_health SET schema_version = 18 WHERE singleton = 1;

-- +goose Down
UPDATE manifest_proposals
SET schema_version = replace(schema_version, 'envarium', 'switchyard'),
    candidate_json = replace(candidate_json, 'envarium', 'switchyard'),
    confidence_json = replace(confidence_json, 'envarium', 'switchyard'),
    unresolved_json = replace(unresolved_json, 'envarium', 'switchyard'),
    validation_json = replace(validation_json, 'envarium', 'switchyard')
WHERE project_id IN (SELECT id FROM projects WHERE slug = 'envarium')
  AND (
      schema_version LIKE '%envarium%'
      OR candidate_json LIKE '%envarium%'
      OR confidence_json LIKE '%envarium%'
      OR unresolved_json LIKE '%envarium%'
      OR validation_json LIKE '%envarium%'
  );

UPDATE manifest_snapshots
SET manifest_json = replace(manifest_json, 'envarium', 'switchyard')
WHERE project_id IN (SELECT id FROM projects WHERE slug = 'envarium')
  AND manifest_json LIKE '%envarium%';

UPDATE settings
SET document_json = replace(
    replace(
        replace(document_json, 'envarium.dev', 'switchyard.dev'),
        '.envarium', '.switchyard'
    ),
    'Envarium', 'Switchyard'
)
WHERE document_json LIKE '%envarium%'
   OR document_json LIKE '%Envarium%';

UPDATE terminal_sessions
SET working_directory = replace(working_directory, '/envarium', '/switchyard')
WHERE working_directory LIKE '%/envarium%';

UPDATE log_segments
SET path = replace(path, '/Envarium/', '/Switchyard/')
WHERE path LIKE '%/Envarium/%';

UPDATE plugin_registrations
SET manifest_path = replace(manifest_path, '/Envarium/', '/Switchyard/'),
    executable_path = replace(executable_path, '/Envarium/', '/Switchyard/'),
    arguments_json = replace(arguments_json, '/Envarium/', '/Switchyard/')
WHERE manifest_path LIKE '%/Envarium/%'
   OR executable_path LIKE '%/Envarium/%'
   OR arguments_json LIKE '%/Envarium/%';

UPDATE project_environments
SET path = replace(path, '/envarium', '/switchyard'),
    compose_project_name = replace(compose_project_name, 'envarium', 'switchyard')
WHERE path LIKE '%/envarium%'
   OR compose_project_name LIKE '%envarium%';

UPDATE discovery_evidence
SET source_path = replace(source_path, '/envarium', '/switchyard'),
    data_json = replace(
        replace(data_json, 'envarium.dev', 'switchyard.dev'),
        '.envarium', '.switchyard'
    ),
    warnings_json = replace(warnings_json, 'envarium.dev', 'switchyard.dev')
WHERE source_path LIKE '%/envarium%'
   OR data_json LIKE '%envarium%'
   OR warnings_json LIKE '%envarium%';

UPDATE manifest_proposals
SET schema_version = replace(schema_version, 'envarium.dev', 'switchyard.dev'),
    candidate_json = replace(
        replace(
            replace(candidate_json, 'envarium.dev', 'switchyard.dev'),
            '.envarium', '.switchyard'
        ),
        'Envarium', 'Switchyard'
    ),
    confidence_json = replace(confidence_json, 'envarium.dev', 'switchyard.dev'),
    unresolved_json = replace(unresolved_json, 'envarium.dev', 'switchyard.dev'),
    validation_json = replace(validation_json, 'envarium.dev', 'switchyard.dev')
WHERE schema_version LIKE '%envarium%'
   OR candidate_json LIKE '%envarium%'
   OR candidate_json LIKE '%Envarium%'
   OR confidence_json LIKE '%envarium%'
   OR unresolved_json LIKE '%envarium%'
   OR validation_json LIKE '%envarium%';

UPDATE manifest_snapshots
SET manifest_json = replace(
    replace(
        replace(manifest_json, 'envarium.dev', 'switchyard.dev'),
        '.envarium', '.switchyard'
    ),
    'Envarium', 'Switchyard'
)
WHERE manifest_json LIKE '%envarium%'
   OR manifest_json LIKE '%Envarium%';

UPDATE project_locations
SET path = replace(path, '/envarium', '/switchyard')
WHERE path LIKE '%/envarium';

UPDATE projects
SET slug = CASE slug WHEN 'envarium' THEN 'switchyard' ELSE slug END,
    display_name = CASE display_name WHEN 'Envarium' THEN 'Switchyard' ELSE display_name END,
    description = replace(description, 'Envarium', 'Switchyard'),
    primary_location = replace(primary_location, '/envarium', '/switchyard')
WHERE slug = 'envarium'
   OR display_name = 'Envarium'
   OR primary_location LIKE '%/envarium';

CREATE TABLE runs_v17 (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    service_id TEXT NOT NULL,
    runtime_driver TEXT NOT NULL,
    origin TEXT NOT NULL CHECK (origin IN ('switchyard', 'external')),
    started_at TEXT NOT NULL,
    ended_at TEXT,
    exit_code INTEGER,
    termination_reason TEXT NOT NULL DEFAULT '',
    identity_fingerprint TEXT NOT NULL,
    restart_count INTEGER NOT NULL DEFAULT 0 CHECK (restart_count >= 0),
    operation_id TEXT
);

CREATE TABLE run_processes_v17 (
    run_id TEXT NOT NULL REFERENCES runs_v17(id) ON DELETE CASCADE,
    pid INTEGER NOT NULL,
    process_group_id INTEGER NOT NULL,
    executable TEXT NOT NULL,
    started_at TEXT NOT NULL,
    working_directory TEXT NOT NULL,
    identity_fingerprint TEXT NOT NULL,
    observed_at TEXT NOT NULL,
    PRIMARY KEY (run_id, pid, started_at)
);

INSERT INTO runs_v17 (
    id, project_id, service_id, runtime_driver, origin, started_at, ended_at,
    exit_code, termination_reason, identity_fingerprint, restart_count, operation_id
)
SELECT
    id, project_id, service_id, runtime_driver,
    CASE origin WHEN 'envarium' THEN 'switchyard' ELSE origin END,
    started_at, ended_at, exit_code, termination_reason, identity_fingerprint,
    restart_count, operation_id
FROM runs;

INSERT INTO run_processes_v17 (
    run_id, pid, process_group_id, executable, started_at, working_directory,
    identity_fingerprint, observed_at
)
SELECT
    run_id, pid, process_group_id, executable, started_at,
    replace(working_directory, '/envarium', '/switchyard'),
    identity_fingerprint, observed_at
FROM run_processes;

DROP TABLE run_processes;
DROP TABLE runs;
ALTER TABLE runs_v17 RENAME TO runs;
ALTER TABLE run_processes_v17 RENAME TO run_processes;

CREATE INDEX runs_project_active_idx ON runs (project_id, ended_at, service_id);
CREATE INDEX run_processes_group_idx ON run_processes (process_group_id, run_id);

UPDATE system_health SET schema_version = 17 WHERE singleton = 1;
