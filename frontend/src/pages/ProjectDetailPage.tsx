import { useParams, Link } from 'react-router-dom';
import { useState } from 'react';
import { useProject, useManifest } from '../hooks';
import { retryProject, downloadVideo } from '../api';
import { useToast } from '../ToastContext';
import StatusBadge from '../components/StatusBadge';
import ProgressTimeline from '../components/ProgressTimeline';
import SceneCard from '../components/SceneCard';
import VideoPlayer from '../components/VideoPlayer';
import ReviewActions from '../components/ReviewActions';
import Skeleton from '../components/Skeleton';

export default function ProjectDetailPage() {
  const { id } = useParams<{ id: string }>();
  const { addToast } = useToast();
  const { project, loading: projectLoading, error: projectError, refetch } = useProject(id);
  const { manifest, loading: manifestLoading, refetch: refetchManifest } = useManifest(id);
  const [retrying, setRetrying] = useState(false);
  const [downloading, setDownloading] = useState(false);
  const [scriptExpanded, setScriptExpanded] = useState(false);

  if (projectLoading) {
    return (
      <div className="page-detail" id="project-detail-page">
        <Skeleton width="120px" height="16px" />
        <div style={{ marginTop: '24px' }}>
          <Skeleton width="60%" height="32px" />
          <div style={{ marginTop: '12px' }}>
            <Skeleton width="100px" height="28px" radius="14px" />
          </div>
        </div>
      </div>
    );
  }

  if (projectError || !project) {
    return (
      <div className="page-detail" id="project-detail-page">
        <Link to="/" className="back-link">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <line x1="19" y1="12" x2="5" y2="12" />
            <polyline points="12 19 5 12 12 5" />
          </svg>
          Back to Dashboard
        </Link>
        <div className="error-state">
          <h2>Project not found</h2>
          <p>{projectError || 'The project could not be loaded.'}</p>
        </div>
      </div>
    );
  }

  const title = project.title || project.topic || 'Untitled Project';
  const date = new Date(project.created_at).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  const showVideo = ['IN_REVIEW', 'RENDER_READY', 'APPROVED'].includes(project.status);
  const showReview = project.status === 'IN_REVIEW';
  const showRetry = project.status === 'FAILED';
  const showDownload = ['IN_REVIEW', 'RENDER_READY', 'APPROVED'].includes(project.status);

  async function handleRetry() {
    if (!id) return;
    setRetrying(true);
    try {
      await retryProject(id);
      addToast('Retrying project...', 'info');
      refetch();
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Retry failed', 'error');
    } finally {
      setRetrying(false);
    }
  }

  async function handleDownload() {
    if (!id) return;
    setDownloading(true);
    try {
      await downloadVideo(id);
      addToast('Download started', 'success');
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Download failed', 'error');
    } finally {
      setDownloading(false);
    }
  }

  function handleReviewAction() {
    refetch();
    refetchManifest();
  }

  const script = manifest?.script;
  const scenes = manifest?.scenes;

  return (
    <div className="page-detail" id="project-detail-page">
      <Link to="/" className="back-link" id="back-to-dashboard">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <line x1="19" y1="12" x2="5" y2="12" />
          <polyline points="12 19 5 12 12 5" />
        </svg>
        Back to Dashboard
      </Link>

      {/* Header */}
      <div className="detail-header">
        <div className="detail-header-top">
          <h1 className="detail-title">{title}</h1>
          <StatusBadge status={project.status} />
        </div>
        {project.topic && project.title && (
          <p className="detail-topic">Topic: {project.topic}</p>
        )}
        <div className="detail-meta">
          <span className="detail-date">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <circle cx="12" cy="12" r="10" />
              <polyline points="12 6 12 12 16 14" />
            </svg>
            {date}
          </span>
          <span className="detail-id">ID: {project.external_id}</span>
        </div>

        {/* Action buttons */}
        <div className="detail-actions">
          {showDownload && (
            <button className="btn btn-secondary" onClick={handleDownload} disabled={downloading} id="btn-download">
              {downloading ? <span className="spinner" /> : (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                  <polyline points="7 10 12 15 17 10" />
                  <line x1="12" y1="15" x2="12" y2="3" />
                </svg>
              )}
              Download
            </button>
          )}
          {showRetry && (
            <button className="btn btn-warning" onClick={handleRetry} disabled={retrying} id="btn-retry">
              {retrying ? <span className="spinner" /> : (
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="23 4 23 10 17 10" />
                  <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10" />
                </svg>
              )}
              Retry
            </button>
          )}
        </div>
      </div>

      {/* Progress Timeline */}
      <section className="detail-section">
        <h2 className="section-title">Pipeline Progress</h2>
        <div className="timeline-container">
          <ProgressTimeline currentStatus={project.status} />
        </div>
      </section>

      {/* Video Player */}
      {showVideo && (
        <section className="detail-section">
          <h2 className="section-title">Video Preview</h2>
          <VideoPlayer projectId={project.project_id} />
        </section>
      )}

      {/* Review Actions */}
      {showReview && (
        <section className="detail-section">
          <ReviewActions projectId={project.project_id} onAction={handleReviewAction} />
        </section>
      )}

      {/* Script */}
      {!manifestLoading && script && (
        <section className="detail-section">
          <button
            className="section-toggle"
            onClick={() => setScriptExpanded(!scriptExpanded)}
            type="button"
          >
            <h2 className="section-title" style={{ margin: 0 }}>Script</h2>
            <svg
              className={`scene-chevron ${scriptExpanded ? 'chevron-open' : ''}`}
              width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor"
              strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
            >
              <polyline points="6 9 12 15 18 9" />
            </svg>
          </button>
          {scriptExpanded && (
            <div className="script-content">
              {script.full_script ? (
                <div className="script-block">
                  <span className="script-label">Full Script</span>
                  <p className="script-text">{script.full_script}</p>
                </div>
              ) : (
                <>
                  {script.hook && (
                    <div className="script-block">
                      <span className="script-label">Hook</span>
                      <p className="script-text">{script.hook}</p>
                    </div>
                  )}
                  {script.setup_text && (
                    <div className="script-block">
                      <span className="script-label">Setup</span>
                      <p className="script-text">{script.setup_text}</p>
                    </div>
                  )}
                  {script.build_text && (
                    <div className="script-block">
                      <span className="script-label">Build</span>
                      <p className="script-text">{script.build_text}</p>
                    </div>
                  )}
                  {script.turning_point_text && (
                    <div className="script-block">
                      <span className="script-label">Turning Point</span>
                      <p className="script-text">{script.turning_point_text}</p>
                    </div>
                  )}
                  {script.collapse_text && (
                    <div className="script-block">
                      <span className="script-label">Collapse</span>
                      <p className="script-text">{script.collapse_text}</p>
                    </div>
                  )}
                  {script.conclusion_text && (
                    <div className="script-block">
                      <span className="script-label">Conclusion</span>
                      <p className="script-text">{script.conclusion_text}</p>
                    </div>
                  )}
                </>
              )}
            </div>
          )}
        </section>
      )}

      {/* Scenes */}
      {!manifestLoading && scenes && scenes.length > 0 && (
        <section className="detail-section">
          <h2 className="section-title">Scenes ({scenes.length})</h2>
          <div className="scenes-list">
            {scenes.map((scene) => (
              <SceneCard key={scene.scene_number} scene={scene} />
            ))}
          </div>
        </section>
      )}
    </div>
  );
}
