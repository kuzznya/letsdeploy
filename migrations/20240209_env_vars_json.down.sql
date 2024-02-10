CREATE TABLE IF NOT EXISTS env_var (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    service_id int NOT NULL REFERENCES service(id) ON DELETE CASCADE,
    name text NOT NULL,
    value text,
    secret_id int REFERENCES secret(id) ON DELETE CASCADE,
    CHECK (value IS NOT NULL AND secret_id IS NULL OR secret_id IS NOT NULL AND value IS NULL),
    UNIQUE (service_id, name)
);

INSERT INTO env_var (service_id, name, value, secret_id)
SELECT s.id, service_env_var->>'name', service_env_var->>'value', nullif(service_env_var->'secret_id', 'null')::int
FROM (
    SELECT service.id, jsonb_array_elements(service.env_vars) service_env_var
    FROM service
) s;

ALTER TABLE service DROP COLUMN IF EXISTS env_var;
