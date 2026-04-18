import { Link } from 'react-router-dom';
import CreateProjectForm from '../components/CreateProjectForm';

export default function CreatePage() {
  return (
    <div className="page-create" id="create-page">
      <Link to="/" className="back-link" id="back-to-dashboard">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <line x1="19" y1="12" x2="5" y2="12" />
          <polyline points="12 19 5 12 12 5" />
        </svg>
        Back to Dashboard
      </Link>
      <div className="page-header">
        <h1 className="page-title">New Project</h1>
        <p className="page-subtitle">Create a new automated video</p>
      </div>
      <CreateProjectForm />
    </div>
  );
}
