import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createProject } from '../api';
import { saveProjectId } from '../storage';
import { useToast } from '../ToastContext';

type Mode = 'topic' | 'script';

export default function CreateProjectForm() {
  const navigate = useNavigate();
  const { addToast } = useToast();
  const [mode, setMode] = useState<Mode>('topic');
  const [loading, setLoading] = useState(false);

  // Form fields
  const [title, setTitle] = useState('');
  const [topic, setTopic] = useState('');
  const [script, setScript] = useState('');
  const [duration, setDuration] = useState(60);
  const [aspectRatio, setAspectRatio] = useState('9:16');
  const [voiceId, setVoiceId] = useState('Matthew');
  const [voiceEngine, setVoiceEngine] = useState('neural');

  const canSubmit = mode === 'topic' ? topic.trim().length > 0 : script.trim().length > 0;

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!canSubmit || loading) return;

    setLoading(true);
    try {
      const payload = mode === 'topic'
        ? {
            topic: topic.trim(),
            ...(title.trim() && { title: title.trim() }),
            channel_style: 'dramatic_history_shorts',
            target_duration_sec: duration,
            aspect_ratio: aspectRatio,
            voice_id: voiceId,
            voice_engine: voiceEngine,
          }
        : {
            title: title.trim() || 'Untitled',
            script: script.trim(),
            channel_style: 'dramatic_history_shorts',
            target_duration_sec: duration,
            aspect_ratio: aspectRatio,
            voice_id: voiceId,
            voice_engine: voiceEngine,
          };

      const project = await createProject(payload);
      saveProjectId(project.project_id);
      addToast('Project created successfully!', 'success');
      navigate(`/projects/${project.project_id}`);
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Failed to create project', 'error');
    } finally {
      setLoading(false);
    }
  }

  return (
    <form className="create-form" onSubmit={handleSubmit} id="create-project-form">
      {/* Mode Toggle */}
      <div className="form-section">
        <label className="form-label">Generation Mode</label>
        <div className="mode-toggle">
          <button
            type="button"
            className={`mode-btn ${mode === 'topic' ? 'mode-active' : ''}`}
            onClick={() => setMode('topic')}
            id="mode-topic"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
              <line x1="12" y1="17" x2="12.01" y2="17" />
            </svg>
            Generate from Topic
          </button>
          <button
            type="button"
            className={`mode-btn ${mode === 'script' ? 'mode-active' : ''}`}
            onClick={() => setMode('script')}
            id="mode-script"
          >
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
              <polyline points="14 2 14 8 20 8" />
              <line x1="16" y1="13" x2="8" y2="13" />
              <line x1="16" y1="17" x2="8" y2="17" />
            </svg>
            Custom Script
          </button>
        </div>
      </div>

      {/* Title */}
      <div className="form-section">
        <label className="form-label" htmlFor="field-title">
          Title
          {mode === 'topic' && <span className="form-optional">Optional</span>}
        </label>
        <input
          id="field-title"
          type="text"
          className="form-input"
          placeholder="Enter video title..."
          value={title}
          onChange={(e) => setTitle(e.target.value)}
        />
      </div>

      {/* Topic (Topic mode) */}
      {mode === 'topic' && (
        <div className="form-section">
          <label className="form-label" htmlFor="field-topic">
            Topic <span className="form-required">*</span>
          </label>
          <input
            id="field-topic"
            type="text"
            className="form-input"
            placeholder="e.g., The Rise of the Zulu Kingdom"
            value={topic}
            onChange={(e) => setTopic(e.target.value)}
            required
          />
          <p className="form-hint">AI will generate a complete script from this topic</p>
        </div>
      )}

      {/* Script (Script mode) */}
      {mode === 'script' && (
        <div className="form-section">
          <label className="form-label" htmlFor="field-script">
            Script <span className="form-required">*</span>
          </label>
          <textarea
            id="field-script"
            className="form-input form-textarea"
            placeholder="Paste your full script here...&#10;&#10;Multiple paragraphs supported."
            value={script}
            onChange={(e) => setScript(e.target.value)}
            rows={8}
            required
          />
          <p className="form-hint">Provide the complete narration script for your video</p>
        </div>
      )}

      {/* Duration & Aspect Ratio */}
      <div className="form-row">
        <div className="form-section">
          <label className="form-label" htmlFor="field-duration">Target Duration</label>
          <select
            id="field-duration"
            className="form-input form-select"
            value={duration}
            onChange={(e) => setDuration(Number(e.target.value))}
          >
            <option value={30}>30 seconds</option>
            <option value={60}>60 seconds</option>
            <option value={90}>90 seconds</option>
            <option value={120}>120 seconds</option>
          </select>
        </div>
        <div className="form-section">
          <label className="form-label" htmlFor="field-aspect">Aspect Ratio</label>
          <select
            id="field-aspect"
            className="form-input form-select"
            value={aspectRatio}
            onChange={(e) => setAspectRatio(e.target.value)}
          >
            <option value="9:16">9:16 Portrait (Shorts)</option>
            <option value="16:9">16:9 Landscape</option>
            <option value="1:1">1:1 Square</option>
          </select>
        </div>
      </div>

      {/* Voice Settings */}
      <div className="form-section">
        <label className="form-label">Voice Settings</label>
        <div className="form-row">
          <div className="form-section">
            <label className="form-label form-label-sm" htmlFor="field-voice-id">Voice</label>
            <select
              id="field-voice-id"
              className="form-input form-select"
              value={voiceId}
              onChange={(e) => setVoiceId(e.target.value)}
            >
              <optgroup label="US English (Male)">
                <option value="Matthew">Matthew - Authoritative</option>
                <option value="Joey">Joey - Casual</option>
                <option value="Justin">Justin - Young</option>
                <option value="Stephen">Stephen - Professional</option>
              </optgroup>
              <optgroup label="US English (Female)">
                <option value="Joanna">Joanna - Warm</option>
                <option value="Kendra">Kendra - Neutral</option>
                <option value="Kimberly">Kimberly - Conversational</option>
                <option value="Salli">Salli - Friendly</option>
                <option value="Ruth">Ruth - News Anchor</option>
                <option value="Ivy">Ivy - Young</option>
              </optgroup>
              <optgroup label="UK English">
                <option value="Brian">Brian - Professional (Male)</option>
                <option value="Amy">Amy - Clear (Female)</option>
                <option value="Emma">Emma - Warm (Female)</option>
                <option value="Arthur">Arthur - Professional (Male)</option>
              </optgroup>
              <optgroup label="Australian English">
                <option value="Russell">Russell - Professional (Male)</option>
                <option value="Nicole">Nicole - Clear (Female)</option>
                <option value="Olivia">Olivia - Warm (Female)</option>
              </optgroup>
              <optgroup label="Other English">
                <option value="Ayanda">Ayanda - South African (Female)</option>
                <option value="Kajal">Kajal - Indian (Female)</option>
              </optgroup>
            </select>
          </div>
          <div className="form-section">
            <label className="form-label form-label-sm" htmlFor="field-voice-engine">Quality</label>
            <select
              id="field-voice-engine"
              className="form-input form-select"
              value={voiceEngine}
              onChange={(e) => setVoiceEngine(e.target.value)}
            >
              <option value="generative">Generative (Most Natural - Short Videos)</option>
              <option value="long-form">Long-form (Best for 3+ Min Videos)</option>
              <option value="neural">Neural (Great Quality - Recommended)</option>
              <option value="standard">Standard (Basic Quality)</option>
            </select>
          </div>
        </div>
        <p className="form-hint">
          {voiceEngine === 'generative' && '🎯 Premium quality - Most natural for videos under 3 minutes. Only available for Ruth, Stephen, Matthew, Joanna.'}
          {voiceEngine === 'long-form' && '📖 Best for extended content (3+ minutes) - Consistent quality throughout. Higher cost but maintains tone.'}
          {voiceEngine === 'neural' && '⭐ Recommended - Natural sounding with great value.'}
          {voiceEngine === 'standard' && '💰 Basic quality - Good for testing.'}
        </p>
      </div>

      {/* Submit */}
      <button
        type="submit"
        className="btn btn-primary btn-lg"
        disabled={!canSubmit || loading}
        id="submit-project"
      >
        {loading ? (
          <>
            <span className="spinner" />
            Creating...
          </>
        ) : (
          <>
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <path d="M22 2L11 13" />
              <path d="M22 2L15 22L11 13L2 9L22 2Z" />
            </svg>
            Create Project
          </>
        )}
      </button>
    </form>
  );
}
