import React from 'react';
import type { Thread } from '../types';
import { ThreadCard } from './ThreadCard';
import { ThreadSkeleton } from './ui/Skeleton';
import { EmptyState } from './EmptyState';

interface ThreadListProps {
  threads: Thread[];
  isLoading?: boolean;
  error?: string | null;
}

export const ThreadList: React.FC<ThreadListProps> = ({ threads, isLoading, error }) => {
  if (isLoading) {
    return (
      <div>
        <ThreadSkeleton />
        <ThreadSkeleton />
        <ThreadSkeleton />
      </div>
    );
  }

  if (error) {
    return (
      <div className="form-error" style={{ padding: '1.5rem 0' }}>
        Failed to load conversations: {error}
      </div>
    );
  }

  if (!threads || threads.length === 0) {
    return <EmptyState />;
  }

  return (
    <div>
      {threads.map((t) => (
        <ThreadCard key={t.id} thread={t} />
      ))}
    </div>
  );
};
