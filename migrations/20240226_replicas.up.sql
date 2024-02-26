ALTER TABLE service ADD COLUMN replicas int NOT NULL DEFAULT 1 CHECK ( replicas >= 0 );
