import { useState } from 'react';
import type { Scene } from '../types';

interface SceneCardProps {
  scene: Scene;
}

export default function SceneCard({ scene }: SceneCardProps) {
  const [expanded, setExpanded] = useState(false);

  return (
    <div className={`scene-card ${expanded ? 'scene-expanded' : ''}`} id={`scene-${scene.scene_number}`}>
      <button
        className="scene-header"
        onClick={() => setExpanded(!expanded)}
        type="button"
        aria-expanded={expanded}
      >
        <div className="scene-header-left">
          <span className="scene-number">Scene {scene.scene_number}</span>
          <span className="scene-duration">{scene.duration_sec.toFixed(1)}s</span>
          {scene.mood && <span className="scene-mood">{scene.mood}</span>}
        </div>
        <svg
          className={`scene-chevron ${expanded ? 'chevron-open' : ''}`}
          width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor"
          strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
        >
          <polyline points="6 9 12 15 18 9" />
        </svg>
      </button>
      {expanded && (
        <div className="scene-body">
          <div className="scene-field">
            <span className="scene-field-label">Narration</span>
            <p className="scene-field-value">{scene.narration_text}</p>
          </div>
          {scene.visual_prompt && (
            <div className="scene-field">
              <span className="scene-field-label">Visual Prompt</span>
              <p className="scene-field-value scene-prompt">{scene.visual_prompt}</p>
            </div>
          )}
          {scene.keywords && scene.keywords.length > 0 && (
            <div className="scene-field">
              <span className="scene-field-label">Keywords</span>
              <div className="scene-keywords">
                {scene.keywords.map((kw) => (
                  <span key={kw} className="keyword-tag">{kw}</span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
