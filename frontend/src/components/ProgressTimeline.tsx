import { STATUS_PIPELINE, getStatusColor, isProcessing, type ProjectStatus } from '../types';

interface ProgressTimelineProps {
  currentStatus: ProjectStatus;
}

export default function ProgressTimeline({ currentStatus }: ProgressTimelineProps) {
  const currentIndex = STATUS_PIPELINE.indexOf(currentStatus);
  const isFailed = currentStatus === 'FAILED';
  const isTerminal = ['APPROVED', 'REJECTED', 'CANCELLED', 'FAILED'].includes(currentStatus);

  return (
    <div className="progress-timeline">
      {STATUS_PIPELINE.map((status, i) => {
        let state: 'done' | 'current' | 'pending' | 'failed';

        if (isFailed && i === currentIndex) {
          state = 'failed';
        } else if (i < currentIndex || (isTerminal && currentIndex === -1 && ['APPROVED', 'REJECTED'].includes(currentStatus))) {
          state = 'done';
        } else if (i === currentIndex) {
          state = 'current';
        } else {
          state = 'pending';
        }

        const color = state === 'done' ? 'green' : state === 'current' ? getStatusColor(status) : state === 'failed' ? 'red' : 'gray';
        const displayName = status.replace(/_/g, ' ');

        return (
          <div key={status} className={`timeline-step timeline-${state}`}>
            <div className="timeline-connector">
              <div className={`timeline-node node-${color}`}>
                {state === 'done' ? (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                ) : state === 'failed' ? (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                    <line x1="18" y1="6" x2="6" y2="18" />
                    <line x1="6" y1="6" x2="18" y2="18" />
                  </svg>
                ) : state === 'current' && isProcessing(status) ? (
                  <span className="timeline-pulse" />
                ) : null}
              </div>
              {i < STATUS_PIPELINE.length - 1 && (
                <div className={`timeline-line line-${state === 'done' ? 'done' : 'pending'}`} />
              )}
            </div>
            <span className={`timeline-label label-${state}`}>{displayName}</span>
          </div>
        );
      })}

      {/* Terminal statuses */}
      {isTerminal && (
        <div className={`timeline-step timeline-current`}>
          <div className="timeline-connector">
            <div className={`timeline-node node-${getStatusColor(currentStatus)}`}>
              {currentStatus === 'APPROVED' ? (
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                  <polyline points="20 6 9 17 4 12" />
                </svg>
              ) : (
                <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="18" y1="6" x2="6" y2="18" />
                  <line x1="6" y1="6" x2="18" y2="18" />
                </svg>
              )}
            </div>
          </div>
          <span className={`timeline-label label-current`}>{currentStatus.replace(/_/g, ' ')}</span>
        </div>
      )}
    </div>
  );
}
