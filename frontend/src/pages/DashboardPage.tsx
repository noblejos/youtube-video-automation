import { useProjects } from '../hooks';
import ProjectCard from '../components/ProjectCard';
import { ProjectCardSkeleton } from '../components/Skeleton';
import { Link } from 'react-router-dom';

export default function DashboardPage() {
  const { projects, loading } = useProjects();

  return (
    <div className="page-dashboard" id="dashboard-page">
      <div className="page-header">
        <div>
          <h1 className="page-title">Projects</h1>
          <p className="page-subtitle">Manage your video automation pipeline</p>
        </div>
      </div>

      {loading ? (
        <div className="projects-grid">
          {Array.from({ length: 3 }).map((_, i) => (
            <ProjectCardSkeleton key={i} />
          ))}
        </div>
      ) : projects.length === 0 ? (
        <div className="empty-state" id="empty-state">
          <div className="empty-icon">
            <svg width="64" height="64" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
              <polygon points="23 7 16 12 23 17 23 7" />
              <rect x="1" y="5" width="15" height="14" rx="2" ry="2" />
            </svg>
          </div>
          <h2 className="empty-title">No projects yet</h2>
          <p className="empty-text">Create your first video project to get started</p>
          <Link to="/create" className="btn btn-primary btn-lg" id="cta-create">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="12" y1="5" x2="12" y2="19" />
              <line x1="5" y1="12" x2="19" y2="12" />
            </svg>
            Create First Project
          </Link>
        </div>
      ) : (
        <div className="projects-grid">
          {projects.map((project) => (
            <ProjectCard key={project.project_id} project={project} />
          ))}
        </div>
      )}
    </div>
  );
}
