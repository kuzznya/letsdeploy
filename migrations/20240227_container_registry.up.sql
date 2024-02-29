CREATE TABLE container_registry (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id text NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    url text NOT NULL,
    username text NOT NULL,
    password text NOT NULL,

    UNIQUE (project_id, url)
);
