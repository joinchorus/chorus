import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchThreadDetail, createMessage } from '../lib/api';
import { ThreadHeader } from '../components/ThreadHeader';
import { Message } from '../components/Message';
import { ReplyForm } from '../components/ReplyForm';
import { ThreadSkeleton } from '../components/ui/Skeleton';
import { EmptyState } from '../components/EmptyState';

export const ThreadDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const queryClient = useQueryClient();
  const [isPosting, setIsPosting] = useState(false);

  const threadId = id || '';

  const { data: detail, isLoading, error } = useQuery({
    queryKey: ['threadDetail', threadId],
    queryFn: () => fetchThreadDetail(threadId),
    enabled: !!threadId,
  });

  const replyMutation = useMutation({
    mutationFn: async ({ body, showFlag }: { body: string; showFlag: boolean }) => {
      const storedName = localStorage.getItem('chorus_conversation_name') || undefined;
      return createMessage(threadId, {
        body,
        show_country: showFlag,
        conversation_name: storedName,
      });
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['threadDetail', threadId] });
    },
  });

  const handleReply = async (body: string, showFlag: boolean) => {
    setIsPosting(true);
    try {
      await replyMutation.mutateAsync({ body, showFlag });
    } finally {
      setIsPosting(false);
    }
  };

  if (isLoading) {
    return (
      <div style={{ width: '100%' }}>
        <ThreadSkeleton />
        <ThreadSkeleton />
      </div>
    );
  }

  if (error || !detail || !detail.thread) {
    return (
      <div className="empty-state" style={{ width: '100%', margin: '2rem 0' }}>
        <h3 className="empty-state-title">Conversation Not Found</h3>
        <p className="empty-state-desc">
          {error instanceof Error ? error.message : 'This conversation may have been removed or does not exist.'}
        </p>
        <Link to="/" className="btn btn-secondary">
          Return to Conversations
        </Link>
      </div>
    );
  }

  const { thread, messages } = detail;

  // Extract unique participant names for this thread
  const participantSet = new Set<string>();
  if (thread.conversation_name) participantSet.add(thread.conversation_name);
  if (messages) {
    messages.forEach((m) => {
      if (m.conversation_name) participantSet.add(m.conversation_name);
    });
  }
  const participantNames = Array.from(participantSet);

  return (
    <div style={{ width: '100%' }}>
      <div style={{ marginBottom: '1.25rem' }}>
        <Link to="/" style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '0.875rem', fontWeight: 600 }}>
          &larr; Back to Conversations
        </Link>
      </div>

      <ThreadHeader thread={thread} participantNames={participantNames} />

      {/* Messages List Dialogue View */}
      <section style={{ marginBottom: '2.5rem' }}>
        <h3 style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-muted)', marginBottom: '1rem' }}>
          Conversation ({messages ? messages.length : 0} replies)
        </h3>

        {messages && messages.length > 0 ? (
          messages.map((msg) => (
            <Message key={msg.id} message={msg} />
          ))
        ) : (
          <EmptyState
            title="No one has replied yet."
            description="Be the first to join this conversation."
          />
        )}
      </section>

      {/* Reply Section */}
      <section>
        <div style={{ fontSize: '0.75rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
          Reply to Conversation
        </div>
        <ReplyForm onSubmit={handleReply} isSubmitting={isPosting} />
      </section>
    </div>
  );
};
