CREATE TABLE IF NOT EXISTS project (
    id varchar(100) PRIMARY KEY GENERATED ALWAYS AS IDENTITY
);

CREATE TABLE project_participant (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id varchar(100) NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    username text NOT NULL,
    UNIQUE(project_id, username)
);

CREATE TABLE service (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id varchar(100) NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name varchar(100) NOT NULL,
    image text NOT NULL,
    port int NOT NULL,
    UNIQUE(project_id, name)
);

CREATE TABLE managed_service (
    id int PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    project_id varchar(100) NOT NULL REFERENCES project(id) ON DELETE CASCADE,
    name text NOT NULL,
    type text NOT NULL,
    UNIQUE(project_id, name)
);
