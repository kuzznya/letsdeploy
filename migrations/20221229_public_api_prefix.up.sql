ALTER TABLE service ADD COLUMN public_api_prefix text;

ALTER TABLE service ADD CONSTRAINT service_project_id_public_api_prefix
    UNIQUE (project_id, public_api_prefix);
