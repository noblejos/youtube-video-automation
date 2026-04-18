import { useNavigate } from 'react-router-dom';
import StatusBadge from './StatusBadge';
import { isProcessing, type Project } from '../types';

interface ProjectCardProps {
  project: Project;
}

export default function ProjectCard({ project }: ProjectCardProps) {
  const navigate = useNavigate();
  const title = project.title || project.topic || 'Untitled Project';
  const processing = isProcessing(project.status);
  const date = new Date(project.created_at).toLocaleDateString('en-US', {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  });

  return (
    <div
      className={`project-card ${processing ? 'card-processing' : ''}`}
      onClick={() => navigate(`/projects/${project.project_id}`)}
      id={`project-${project.external_id}`}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => e.key === 'Enter' && navigate(`/projects/${project.project_id}`)}
    >
      <div className="card-header">
        <h3 className="card-title">{title}</h3>
        <StatusBadge status={project.status} size="sm" />
      </div>
      {project.topic && project.title && (
        <p className="card-topic">{project.topic}</p>
      )}
      <div className="card-footer">
        <span className="card-date">{date}</span>
        <span className="card-id">{project.external_id}</span>
      </div>
      {processing && (
        <div className="card-progress-bar">
          <div className="card-progress-fill" />
        </div>
      )}
    </div>
  );
}
