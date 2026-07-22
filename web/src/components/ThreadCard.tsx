import React from 'react';
import { Link } from 'react-router-dom';
import type { Thread } from '../types';

interface ThreadCardProps {
  thread: Thread;
}

export const ThreadCard: React.FC<ThreadCardProps> = ({ thread }) => {
  const formattedDate = new Date(thread.created_at).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <div
      style={{
        borderBottom: '1px solid #e1e4e8',
        paddingTop: '1rem',
        paddingBottom: '1rem',
      }}
    >
      <h3 style={{ fontSize: '1.0625rem', fontWeight: 600, marginBottom: '0.25rem' }}>
        <Link to={`/thread/${thread.id}`} style={{ color: '#0969da' }}>
          {thread.title}
        </Link>
      </h3>
      <div style={{ fontSize: '0.8125rem', color: '#6b7280' }}>
        posted by <span className="font-mono" style={{ color: '#24292f' }}>{thread.author_id}</span>
        {thread.country && <span title={`Country: ${thread.country}`} style={{ marginLeft: '0.375rem' }}>[{thread.country}]</span>}
        &bull; {formattedDate}
      </div>
    </div>
  );
};
