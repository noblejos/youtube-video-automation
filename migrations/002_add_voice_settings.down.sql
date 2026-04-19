-- Remove voice settings from projects table
ALTER TABLE projects DROP COLUMN IF EXISTS voice_id;
ALTER TABLE projects DROP COLUMN IF EXISTS voice_engine;
