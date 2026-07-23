import React from 'react';
import { Link } from 'react-router-dom';

interface EmptyStateProps {
  title?: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'No conversations yet.',
  description = 'Start the first thoughtful discussion.',
  actionLabel = 'Start Conversation',
  onAction,
}) => {
  return (
    <div className="editorial-empty-state">
      <h3 className="empty-title">{title}</h3>
      <p className="empty-subtitle">{description}</p>
      {onAction ? (
        <button onClick={onAction} className="btn-editorial-primary">
          {actionLabel}
        </button>
      ) : (
        <Link to="/new" className="btn-editorial-primary">
          {actionLabel}
        </Link>
      )}
    </div>
  );
};
