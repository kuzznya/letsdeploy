CREATE TABLE project (
    id text PRIMARY KEY
);

CREATE TABLE project_participant (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id text NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    username text NOT NULL,
    UNIQUE(project_id, username)
);

CREATE TABLE service (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id text NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name text NOT NULL,
    image text NOT NULL,
    port int NOT NULL,
    UNIQUE(project_id, name)
);

CREATE TABLE managed_service (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id text NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name text NOT NULL,
    type text NOT NULL,
    UNIQUE(project_id, name)
);

CREATE TABLE secret (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id text NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name text NOT NULL,
    value text NOT NULL,
    managed_service_id int REFERENCES managed_service(id) ON DELETE CASCADE,
    UNIQUE (project_id, name)
);

CREATE TABLE env_var (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    service_id int NOT NULL REFERENCES service(id) ON DELETE CASCADE,
    name text NOT NULL,
    value text,
    secret_id int REFERENCES secret(id),
    CHECK (value IS NOT NULL AND secret_id IS NULL OR secret_id IS NOT NULL AND value IS NULL),
    UNIQUE (service_id, name)
);
