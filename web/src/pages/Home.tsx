import React from 'react';
import { useQuery } from '@tanstack/react-query';
import { fetchThreads } from '../lib/api';
import { ThreadList } from '../components/ThreadList';

export const Home: React.FC = () => {
  const { data: threads, isLoading, error } = useQuery({
    queryKey: ['threads'],
    queryFn: fetchThreads,
  });

  return (
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: '1rem',
          borderBottom: '1px solid #e1e4e8',
          paddingBottom: '0.75rem',
        }}
      >
        <h2 style={{ fontSize: '1.25rem', fontWeight: 600, color: '#111827' }}>
          Recent Threads
        </h2>
        <span style={{ fontSize: '0.8125rem', color: '#6b7280' }}>
          {threads ? `${threads.length} discussions` : ''}
        </span>
      </div>

      <ThreadList
        threads={threads || []}
        isLoading={isLoading}
        error={error instanceof Error ? error.message : null}
      />
    </div>
  );
};
