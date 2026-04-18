-- Initial schema for YouTube video automation system

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Projects table
CREATE TABLE projects (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    external_id VARCHAR(255) UNIQUE NOT NULL,
    topic TEXT NOT NULL,
    title TEXT,
    channel_style VARCHAR(100) NOT NULL DEFAULT 'dramatic_history_shorts',
    target_duration_sec INTEGER NOT NULL DEFAULT 120,
    aspect_ratio VARCHAR(20) NOT NULL DEFAULT '9:16',
    status VARCHAR(50) NOT NULL DEFAULT 'CREATED',
    review_required BOOLEAN NOT NULL DEFAULT true,
    current_step VARCHAR(100),
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_status ON projects(status);
CREATE INDEX idx_projects_external_id ON projects(external_id);

-- Scripts table
CREATE TABLE scripts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    hook TEXT,
    setup_text TEXT,
    build_text TEXT,
    turning_point_text TEXT,
    collapse_text TEXT,
    conclusion_text TEXT,
    full_script TEXT,
    raw_model_response JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_scripts_project_id ON scripts(project_id);

-- Scenes table
CREATE TABLE scenes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scene_number INTEGER NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    narration_text TEXT NOT NULL,
    ssml_text TEXT,
    duration_sec NUMERIC(6,2) NOT NULL,
    mood VARCHAR(100),
    keywords JSONB,
    visual_prompt TEXT,
    asset_strategy VARCHAR(50) DEFAULT 'ai_or_archive',
    transition_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    UNIQUE(project_id, scene_number)
);

CREATE INDEX idx_scenes_project_id ON scenes(project_id);
CREATE INDEX idx_scenes_status ON scenes(status);

-- Assets table
CREATE TABLE assets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scene_id UUID REFERENCES scenes(id) ON DELETE CASCADE,
    asset_type VARCHAR(50) NOT NULL,
    provider VARCHAR(100) NOT NULL,
    storage_key TEXT NOT NULL,
    mime_type VARCHAR(100),
    source_url TEXT,
    prompt_used TEXT,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assets_project_id ON assets(project_id);
CREATE INDEX idx_assets_scene_id ON assets(scene_id);

-- Audio files table
CREATE TABLE audio_files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    scene_id UUID NOT NULL REFERENCES scenes(id) ON DELETE CASCADE,
    provider VARCHAR(100) NOT NULL,
    voice_id VARCHAR(100) NOT NULL,
    engine VARCHAR(50),
    storage_key TEXT NOT NULL,
    duration_ms INTEGER,
    speech_marks_key TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audio_files_project_id ON audio_files(project_id);
CREATE INDEX idx_audio_files_scene_id ON audio_files(scene_id);

-- Subtitles table
CREATE TABLE subtitles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    format VARCHAR(20) NOT NULL DEFAULT 'srt',
    storage_key TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subtitles_project_id ON subtitles(project_id);

-- Renders table
CREATE TABLE renders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    render_type VARCHAR(50) NOT NULL,
    storage_key TEXT NOT NULL,
    metadata JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_renders_project_id ON renders(project_id);

-- Jobs table
CREATE TABLE jobs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    job_type VARCHAR(100) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'QUEUED',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    max_attempts INTEGER NOT NULL DEFAULT 3,
    payload JSONB,
    result JSONB,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_jobs_project_id ON jobs(project_id);
CREATE INDEX idx_jobs_status ON jobs(status);
CREATE INDEX idx_jobs_job_type ON jobs(job_type);

-- Review actions table
CREATE TABLE review_actions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    notes TEXT,
    acted_by VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_review_actions_project_id ON review_actions(project_id);

-- Updated_at trigger function
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Apply updated_at triggers
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scripts_updated_at BEFORE UPDATE ON scripts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_scenes_updated_at BEFORE UPDATE ON scenes
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_jobs_updated_at BEFORE UPDATE ON jobs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
