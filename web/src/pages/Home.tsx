import React, { useState, useMemo } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchThreads } from '../lib/api';
import { ThreadCard } from '../components/ThreadCard';
import { ThreadSkeleton } from '../components/ui/Skeleton';
import { EmptyState } from '../components/EmptyState';
import { OFFICIAL_TOPICS } from '../lib/topics';

export const Home: React.FC = () => {
  const [selectedTopic, setSelectedTopic] = useState<string>('all');
  const [searchQuery, setSearchQuery] = useState<string>('');

  const { data: threads, isLoading, error } = useQuery({
    queryKey: ['threads'],
    queryFn: fetchThreads,
  });

  const filteredThreads = useMemo(() => {
    if (!threads) return [];
    return threads.filter((th) => {
      const matchTopic = selectedTopic === 'all' || (th.topic || 'technology') === selectedTopic;
      const q = searchQuery.toLowerCase().trim();
      const matchQuery = !q || 
        th.title.toLowerCase().includes(q) || 
        (th.topic || '').toLowerCase().includes(q) || 
        (th.body || '').toLowerCase().includes(q) ||
        (th.conversation_name || '').toLowerCase().includes(q);
      return matchTopic && matchQuery;
    });
  }, [threads, selectedTopic, searchQuery]);

  return (
    <div style={{ width: '100%', padding: '1.5rem 0' }}>
      
      {/* 1. Header & Primary Action: Start Conversation */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1.75rem', gap: '1rem', flexWrap: 'wrap' }}>
        <div>
          <h1 style={{ fontSize: '2rem', fontWeight: 800, letterSpacing: '-0.04em', color: 'var(--text-primary)', margin: 0, lineHeight: 1.15 }}>
            Recent Conversations
          </h1>
          <p style={{ color: 'var(--text-secondary)', fontSize: '0.9375rem', margin: '0.35rem 0 0', lineHeight: 1.4 }}>
            Thoughtful discourse where identity belongs exclusively to the conversation.
          </p>
        </div>

        <Link
          to="/new"
          style={{
            background: 'var(--btn-primary-bg)',
            color: 'var(--btn-primary-text)',
            fontSize: '0.9375rem',
            fontWeight: 700,
            padding: '0.65rem 1.35rem',
            borderRadius: '8px',
            textDecoration: 'none',
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.4rem',
            transition: 'opacity 0.15s ease',
            whiteSpace: 'nowrap',
          }}
        >
          + Start Conversation
        </Link>
      </div>

      {/* 2. Utility Search & Topic Filters */}
      <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem', marginBottom: '2.5rem' }}>
        {/* Search Input Utility */}
        <div style={{ position: 'relative', width: '100%' }}>
          <input
            type="text"
            placeholder="Search conversations..."
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            style={{
              width: '100%',
              background: 'var(--bg-surface)',
              border: '1px solid var(--border-default)',
              borderRadius: '8px',
              padding: '0.75rem 1.25rem 0.75rem 2.6rem',
              color: 'var(--text-primary)',
              fontSize: '0.9375rem',
              outline: 'none',
              transition: 'border-color 0.15s ease',
            }}
          />
          <svg
            style={{ position: 'absolute', left: '0.9rem', top: '50%', transform: 'translateY(-50%)', color: 'var(--text-muted)' }}
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            strokeWidth="2"
          >
            <circle cx="11" cy="11" r="8"></circle>
            <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
          </svg>
        </div>

        {/* Lightweight Horizontal Topic Filter Chips */}
        <div
          style={{
            display: 'flex',
            gap: '0.4rem',
            overflowX: 'auto',
            paddingBottom: '0.25rem',
            WebkitOverflowScrolling: 'touch',
          }}
        >
          <button
            onClick={() => setSelectedTopic('all')}
            style={{
              padding: '0.35rem 0.85rem',
              borderRadius: '20px',
              fontSize: '0.8125rem',
              fontWeight: 600,
              border: '1px solid var(--border-default)',
              cursor: 'pointer',
              whiteSpace: 'nowrap',
              background: selectedTopic === 'all' ? 'var(--btn-primary-bg)' : 'var(--bg-surface)',
              color: selectedTopic === 'all' ? 'var(--btn-primary-text)' : 'var(--text-secondary)',
              transition: 'all 0.15s ease',
            }}
          >
            All
          </button>

          {OFFICIAL_TOPICS.map((t) => {
            const isSelected = selectedTopic === t.id;
            return (
              <button
                key={t.id}
                onClick={() => setSelectedTopic(t.id)}
                style={{
                  padding: '0.35rem 0.85rem',
                  borderRadius: '20px',
                  fontSize: '0.8125rem',
                  fontWeight: 600,
                  border: '1px solid var(--border-default)',
                  cursor: 'pointer',
                  whiteSpace: 'nowrap',
                  background: isSelected ? 'var(--btn-primary-bg)' : 'var(--bg-surface)',
                  color: isSelected ? 'var(--btn-primary-text)' : 'var(--text-secondary)',
                  transition: 'all 0.15s ease',
                }}
              >
                {t.name}
              </button>
            );
          })}
        </div>
      </div>

      {/* 3. Primary Conversation Cards Feed */}
      {isLoading ? (
        <div>
          <ThreadSkeleton />
          <ThreadSkeleton />
          <ThreadSkeleton />
        </div>
      ) : error ? (
        <div style={{ padding: '2rem', textAlign: 'center', color: 'var(--accent-red)', fontSize: '0.875rem' }}>
          {error instanceof Error ? error.message : 'Failed to load conversations.'}
        </div>
      ) : filteredThreads.length === 0 ? (
        <EmptyState
          title="No conversations found."
          description="Start a new conversation or choose another topic filter."
          actionLabel="Start a Conversation"
          onAction={() => window.location.href = '/new'}
        />
      ) : (
        <div>
          {filteredThreads.map((thread) => (
            <ThreadCard key={thread.id} thread={thread} />
          ))}
        </div>
      )}
    </div>
  );
};
