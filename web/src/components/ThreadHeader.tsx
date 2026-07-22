import React from 'react';
import type { Thread } from '../types';

interface ThreadHeaderProps {
  thread: Thread;
}

export const ThreadHeader: React.FC<ThreadHeaderProps> = ({ thread }) => {
  const formattedDate = new Date(thread.created_at).toLocaleString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div
      style={{
        borderBottom: '1px solid #e1e4e8',
        paddingBottom: '1.25rem',
        marginBottom: '1.5rem',
      }}
    >
      <h1
        style={{
          fontSize: '1.5rem',
          fontWeight: 700,
          color: '#111827',
          marginBottom: '0.5rem',
          lineHeight: '1.3',
        }}
      >
        {thread.title}
      </h1>
      <div style={{ fontSize: '0.8125rem', color: '#6b7280', display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
        <span>
          Thread identity: <span className="font-mono" style={{ color: '#24292f' }}>{thread.author_id}</span>
        </span>
        {thread.country && <span title={`Country: ${thread.country}`}>[{thread.country}]</span>}
        <span>&bull;</span>
        <span>{formattedDate}</span>
      </div>
    </div>
  );
};
