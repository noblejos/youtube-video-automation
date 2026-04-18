import { useState } from 'react';
import { approveProject, rejectProject } from '../api';
import { useToast } from '../ToastContext';

interface ReviewActionsProps {
  projectId: string;
  onAction: () => void;
}

export default function ReviewActions({ projectId, onAction }: ReviewActionsProps) {
  const { addToast } = useToast();
  const [notes, setNotes] = useState('');
  const [loading, setLoading] = useState<'approve' | 'reject' | null>(null);

  async function handleApprove() {
    setLoading('approve');
    try {
      await approveProject(projectId, {
        notes: notes.trim() || 'Approved',
        acted_by: 'dashboard-user',
      });
      addToast('Project approved!', 'success');
      onAction();
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Failed to approve', 'error');
    } finally {
      setLoading(null);
    }
  }

  async function handleReject() {
    setLoading('reject');
    try {
      await rejectProject(projectId, {
        notes: notes.trim() || 'Rejected',
        acted_by: 'dashboard-user',
      });
      addToast('Project rejected', 'info');
      onAction();
    } catch (err) {
      addToast(err instanceof Error ? err.message : 'Failed to reject', 'error');
    } finally {
      setLoading(null);
    }
  }

  return (
    <div className="review-actions" id="review-actions">
      <h3 className="review-title">Review Decision</h3>
      <textarea
        className="form-input form-textarea"
        placeholder="Add notes (optional)..."
        value={notes}
        onChange={(e) => setNotes(e.target.value)}
        rows={3}
        id="review-notes"
      />
      <div className="review-buttons">
        <button
          className="btn btn-success"
          onClick={handleApprove}
          disabled={loading !== null}
          id="btn-approve"
        >
          {loading === 'approve' ? (
            <><span className="spinner" /> Approving...</>
          ) : (
            <>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="20 6 9 17 4 12" />
              </svg>
              Approve
            </>
          )}
        </button>
        <button
          className="btn btn-danger"
          onClick={handleReject}
          disabled={loading !== null}
          id="btn-reject"
        >
          {loading === 'reject' ? (
            <><span className="spinner" /> Rejecting...</>
          ) : (
            <>
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="18" y1="6" x2="6" y2="18" />
                <line x1="6" y1="6" x2="18" y2="18" />
              </svg>
              Reject
            </>
          )}
        </button>
      </div>
    </div>
  );
}
