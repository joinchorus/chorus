import React, { useState, useMemo } from 'react';
import { useSearchParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchThreads, SYSTEM_BOARDS } from '../lib/api';
import { ThreadCard } from '../components/ThreadCard';
import { ThreadSkeleton } from '../components/ui/Skeleton';
import { EmptyState } from '../components/EmptyState';

export const Home: React.FC = () => {
  const [selectedBoardSlug, setSelectedBoardSlug] = useState<string>('all');
  const [searchParams] = useSearchParams();
  const searchQuery = searchParams.get('q') || '';

  const { data: threads, isLoading, error } = useQuery({
    queryKey: ['threads'],
    queryFn: fetchThreads,
  });

  const filteredThreads = useMemo(() => {
    if (!threads) return [];
    return threads.filter((th) => {
      const thBoard = (th.board_slug || th.topic || '').toLowerCase();
      const matchBoard =
        selectedBoardSlug === 'all' ||
        thBoard === selectedBoardSlug.toLowerCase();
      const q = searchQuery.toLowerCase().trim();
      const matchQuery =
        !q ||
        th.title.toLowerCase().includes(q) ||
        (th.topic || '').toLowerCase().includes(q) ||
        (th.board_display_name || '').toLowerCase().includes(q) ||
        (th.body || '').toLowerCase().includes(q) ||
        (th.conversation_name || '').toLowerCase().includes(q);
      return matchBoard && matchQuery;
    });
  }, [threads, selectedBoardSlug, searchQuery]);

  return (
    <div className="home-editorial-wrapper">
      {/* Hero Section */}
      <section className="editorial-hero">
        <h1 className="editorial-hero-title">
          Identity belongs to the conversation.
        </h1>
        <p className="editorial-hero-subtitle">
          A quiet, anonymous space where ideas matter more than profiles, algorithms, or reputation.
        </p>
      </section>

      {/* Lightweight Board Pill Filters */}
      <nav className="topic-filters" aria-label="Board filters">
        <button
          onClick={() => setSelectedBoardSlug('all')}
          className={`topic-pill ${selectedBoardSlug === 'all' ? 'active' : ''}`}
        >
          All
        </button>
        {SYSTEM_BOARDS.map((b) => {
          const isSelected = selectedBoardSlug === b.slug;
          return (
            <button
              key={b.id}
              onClick={() => setSelectedBoardSlug(b.slug)}
              className={`topic-pill ${isSelected ? 'active' : ''}`}
              title={b.description}
            >
              {b.display_name}
            </button>
          );
        })}
      </nav>

      {/* Active Board Context Hint */}
      {selectedBoardSlug !== 'all' && (
        <div style={{ marginBottom: '1.5rem', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
          <span style={{ fontSize: '0.875rem', color: 'var(--text-secondary)' }}>
            Showing conversations in{' '}
            <strong style={{ color: 'var(--text-primary)' }}>
              {SYSTEM_BOARDS.find((b) => b.slug === selectedBoardSlug)?.display_name}
            </strong>
          </span>
          <Link
            to={`/board/${selectedBoardSlug}`}
            style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', textDecoration: 'none' }}
          >
            View Board Context →
          </Link>
        </div>
      )}

      {/* Conversation Feed */}
      <section className="conversation-feed">
        {isLoading ? (
          <div className="skeleton-feed">
            <ThreadSkeleton />
            <ThreadSkeleton />
            <ThreadSkeleton />
          </div>
        ) : error ? (
          <div className="error-feed">
            {error instanceof Error ? error.message : 'Failed to load conversations.'}
          </div>
        ) : filteredThreads.length === 0 ? (
          <EmptyState
            title="No conversations yet."
            description="Start the first thoughtful discussion."
            actionLabel="Start Conversation"
          />
        ) : (
          <div className="feed-cards-list">
            {filteredThreads.map((thread) => (
              <ThreadCard key={thread.id} thread={thread} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
};
