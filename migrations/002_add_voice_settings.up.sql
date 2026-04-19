-- Add voice settings to projects table
ALTER TABLE projects ADD COLUMN voice_id VARCHAR(100) DEFAULT 'Ayanda';
ALTER TABLE projects ADD COLUMN voice_engine VARCHAR(50) DEFAULT 'standard';
