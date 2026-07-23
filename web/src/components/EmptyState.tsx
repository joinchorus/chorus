import React from 'react';
import { Button } from './ui/Button';

interface EmptyStateProps {
  title?: string;
  description?: string;
  actionLabel?: string;
  onAction?: () => void;
}

export const EmptyState: React.FC<EmptyStateProps> = ({
  title = 'No one has started a conversation here yet.',
  description = 'Be the first to open a discussion in this topic container.',
  actionLabel = 'Be the first',
  onAction,
}) => {
  return (
    <div className="empty-state" style={{ padding: '4rem 2rem', textAlign: 'center', background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '12px', margin: '2rem 0' }}>
      <h3 className="empty-state-title" style={{ fontSize: '1.25rem', fontWeight: 700, color: 'var(--text-primary)', marginBottom: '0.5rem' }}>{title}</h3>
      <p className="empty-state-desc" style={{ color: 'var(--text-secondary)', fontSize: '0.9375rem', marginBottom: '1.5rem' }}>{description}</p>
      {onAction && actionLabel && (
        <Button onClick={onAction} size="md">
          {actionLabel}
        </Button>
      )}
    </div>
  );
};
