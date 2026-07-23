import React, { useMemo } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { fetchThreads, SYSTEM_BOARDS } from '../lib/api';
import { ThreadCard } from '../components/ThreadCard';
import { ThreadSkeleton } from '../components/ui/Skeleton';
import { EmptyState } from '../components/EmptyState';

export const BoardDetail: React.FC = () => {
  const { slug } = useParams<{ slug: string }>();

  const currentBoard = useMemo(() => {
    return SYSTEM_BOARDS.find((b) => b.slug.toLowerCase() === (slug || '').toLowerCase());
  }, [slug]);

  const { data: threads, isLoading, error } = useQuery({
    queryKey: ['threads'],
    queryFn: fetchThreads,
  });

  const boardThreads = useMemo(() => {
    if (!threads || !slug) return [];
    const targetSlug = slug.toLowerCase();
    return threads.filter((th) => {
      const thBoardSlug = (th.board_slug || '').toLowerCase();
      const thTopic = (th.topic || '').toLowerCase();
      return thBoardSlug === targetSlug || thTopic === targetSlug;
    });
  }, [threads, slug]);

  if (!currentBoard) {
    return (
      <div className="board-page-wrapper">
        <div className="editorial-empty-state">
          <h2 className="empty-title">Board Not Found</h2>
          <p className="empty-subtitle">The requested board context does not exist in Chorus.</p>
          <Link to="/" className="btn-editorial-primary">
            Return to All Conversations
          </Link>
        </div>
      </div>
    );
  }

  return (
    <div className="board-page-wrapper">
      {/* Board Header Header */}
      <header className="board-header">
        <div className="board-breadcrumb">
          <Link to="/" className="breadcrumb-link">
            ← All Boards
          </Link>
        </div>
        <h1 className="board-title">{currentBoard.display_name}</h1>
        <p className="board-description">{currentBoard.description}</p>
      </header>

      {/* Board Conversations Feed */}
      <section className="conversation-feed">
        {isLoading ? (
          <div className="skeleton-feed">
            <ThreadSkeleton />
            <ThreadSkeleton />
          </div>
        ) : error ? (
          <div className="error-feed">
            {error instanceof Error ? error.message : 'Failed to load board conversations.'}
          </div>
        ) : boardThreads.length === 0 ? (
          <EmptyState
            title={`No conversations in ${currentBoard.display_name} yet.`}
            description="Start the first thoughtful discussion in this context."
            actionLabel="Start Conversation"
          />
        ) : (
          <div className="feed-cards-list">
            {boardThreads.map((thread) => (
              <ThreadCard key={thread.id} thread={thread} />
            ))}
          </div>
        )}
      </section>
    </div>
  );
};
