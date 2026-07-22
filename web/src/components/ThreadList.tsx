import React from 'react';
import type { Thread } from '../types';
import { ThreadCard } from './ThreadCard';

interface ThreadListProps {
  threads: Thread[];
  isLoading?: boolean;
  error?: string | null;
}

export const ThreadList: React.FC<ThreadListProps> = ({ threads, isLoading, error }) => {
  if (isLoading) {
    return (
      <div style={{ padding: '2rem 0', color: '#6b7280', fontSize: '0.875rem' }}>
        Loading threads...
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: '1.5rem 0', color: '#cf222e', fontSize: '0.875rem' }}>
        Failed to load threads: {error}
      </div>
    );
  }

  if (!threads || threads.length === 0) {
    return (
      <div style={{ padding: '2.5rem 0', color: '#6b7280', fontSize: '0.9375rem' }}>
        No threads yet.
      </div>
    );
  }

  return (
    <div>
      {threads.map((t) => (
        <ThreadCard key={t.id} thread={t} />
      ))}
    </div>
  );
};
