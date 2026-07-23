import React, { useState, useMemo } from 'react';
import { useSearchParams } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchThreads } from '../lib/api';
import { ThreadCard } from '../components/ThreadCard';
import { ThreadSkeleton } from '../components/ui/Skeleton';
import { EmptyState } from '../components/EmptyState';

const FILTER_TOPICS = [
  { id: 'all', name: 'All' },
  { id: 'technology', name: 'Technology' },
  { id: 'programming', name: 'Programming' },
  { id: 'ai', name: 'AI' },
  { id: 'science', name: 'Science' },
  { id: 'design', name: 'Design' },
  { id: 'philosophy', name: 'Philosophy' },
  { id: 'history', name: 'History' },
  { id: 'politics', name: 'Politics' },
  { id: 'books', name: 'Books' },
  { id: 'movies', name: 'Movies' },
  { id: 'music', name: 'Music' },
];

export const Home: React.FC = () => {
  const [selectedTopic, setSelectedTopic] = useState<string>('all');
  const [searchParams] = useSearchParams();
  const searchQuery = searchParams.get('q') || '';

  const { data: threads, isLoading, error } = useQuery({
    queryKey: ['threads'],
    queryFn: fetchThreads,
  });

  const filteredThreads = useMemo(() => {
    if (!threads) return [];
    return threads.filter((th) => {
      const matchTopic =
        selectedTopic === 'all' ||
        (th.topic || 'technology').toLowerCase() === selectedTopic.toLowerCase();
      const q = searchQuery.toLowerCase().trim();
      const matchQuery =
        !q ||
        th.title.toLowerCase().includes(q) ||
        (th.topic || '').toLowerCase().includes(q) ||
        (th.body || '').toLowerCase().includes(q) ||
        (th.conversation_name || '').toLowerCase().includes(q);
      return matchTopic && matchQuery;
    });
  }, [threads, selectedTopic, searchQuery]);

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

      {/* Topic Filters */}
      <nav className="topic-filters" aria-label="Topic filters">
        {FILTER_TOPICS.map((topic) => {
          const isSelected = selectedTopic === topic.id;
          return (
            <button
              key={topic.id}
              onClick={() => setSelectedTopic(topic.id)}
              className={`topic-pill ${isSelected ? 'active' : ''}`}
            >
              {topic.name}
            </button>
          );
        })}
      </nav>

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
