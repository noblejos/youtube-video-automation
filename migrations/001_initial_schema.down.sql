-- Rollback initial schema

DROP TRIGGER IF EXISTS update_jobs_updated_at ON jobs;
DROP TRIGGER IF EXISTS update_scenes_updated_at ON scenes;
DROP TRIGGER IF EXISTS update_scripts_updated_at ON scripts;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;

DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS review_actions;
DROP TABLE IF EXISTS renders;
DROP TABLE IF EXISTS subtitles;
DROP TABLE IF EXISTS audio_files;
DROP TABLE IF EXISTS assets;
DROP TABLE IF EXISTS scenes;
DROP TABLE IF EXISTS scripts;
DROP TABLE IF EXISTS jobs;
DROP TABLE IF EXISTS projects;

DROP EXTENSION IF EXISTS "uuid-ossp";
