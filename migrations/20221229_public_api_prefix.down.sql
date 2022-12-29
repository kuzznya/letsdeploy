ALTER TABLE service DROP CONSTRAINT IF EXISTS service_project_id_public_api_prefix;

ALTER TABLE service DROP COLUMN IF EXISTS public_api_prefix;
