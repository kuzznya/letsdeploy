CREATE TABLE IF NOT EXISTS project (
    id INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name TEXT NOT NULL
);

CREATE TABLE project_participant (
    id INT GENERATED ALWAYS AS IDENTITY,
    project_id INT REFERENCES project(id) ON DELETE CASCADE NOT NULL,
    username TEXT NOT NULL,
    UNIQUE(project_id, username)
);
