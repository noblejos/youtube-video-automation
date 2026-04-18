import { type ProjectStatus, getStatusColor, STATUS_DESCRIPTIONS } from '../types';

interface StatusBadgeProps {
  status: ProjectStatus;
  size?: 'sm' | 'md';
}

export default function StatusBadge({ status, size = 'md' }: StatusBadgeProps) {
  const color = getStatusColor(status);
  const label = STATUS_DESCRIPTIONS[status] || status;
  const displayText = status.replace(/_/g, ' ');

  return (
    <span
      className={`status-badge status-${color} status-${size}`}
      title={label}
    >
      <span className="status-dot" />
      {displayText}
    </span>
  );
}
