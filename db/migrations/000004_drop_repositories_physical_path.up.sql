ALTER TABLE repositories DROP CONSTRAINT IF EXISTS repositories_physical_path_key;
ALTER TABLE repositories DROP COLUMN IF EXISTS physical_path;
