ALTER TABLE service ADD COLUMN env_vars jsonb NOT NULL DEFAULT '[]'::jsonb;

UPDATE service SET env_vars = (
    WITH env_vars AS (
        SELECT
            service_id,
            jsonb_set(
                row_to_json(env_var)::jsonb,
                '{secret}',
                coalesce((SELECT '"' || secret.name || '"' FROM secret WHERE secret.id = env_var.secret_id)::jsonb, 'null'::jsonb)
            )::jsonb  - 'id' - 'service_id' - 'secret_id' as new_env_var
        FROM env_var
    )
    SELECT coalesce(
        jsonb_agg(env_vars.new_env_var),
        '[]'::jsonb
    )
    FROM env_vars
    GROUP BY service_id
);

DROP TABLE env_var;
