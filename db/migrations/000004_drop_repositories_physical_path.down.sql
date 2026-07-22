ALTER TABLE repositories ADD COLUMN physical_path VARCHAR NOT NULL DEFAULT '';
ALTER TABLE repositories ADD CONSTRAINT repositories_physical_path_key UNIQUE (physical_path);
