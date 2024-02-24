CREATE TABLE api_key (
    id text PRIMARY KEY,
    username text NOT NULL,
    name text NOT NULL,
    UNIQUE (username, name)
);
