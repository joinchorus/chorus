import React, { useState } from 'react';
import { useParams, Link } from 'react-router-dom';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { fetchThreadDetail, createMessage } from '../lib/api';
import { ThreadHeader } from '../components/ThreadHeader';
import { Message } from '../components/Message';
import { ReplyForm } from '../components/ReplyForm';

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
      return createMessage(threadId, {
        body,
        show_country: showFlag,
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
      <div style={{ padding: '2rem 0', color: '#6b7280', fontSize: '0.875rem' }}>
        Loading thread...
      </div>
    );
  }

  if (error || !detail || !detail.thread) {
    return (
      <div style={{ padding: '2rem 0' }}>
        <p style={{ color: '#cf222e', marginBottom: '1rem' }}>
          {error instanceof Error ? error.message : 'Thread not found or failed to load.'}
        </p>
        <Link to="/" style={{ fontSize: '0.875rem' }}>
          &larr; Return to threads
        </Link>
      </div>
    );
  }

  const { thread, messages } = detail;

  return (
    <div>
      <div style={{ marginBottom: '1rem' }}>
        <Link to="/" style={{ fontSize: '0.8125rem', color: '#6b7280' }}>
          &larr; Back to threads
        </Link>
      </div>

      <ThreadHeader thread={thread} />

      {/* Messages List */}
      <section style={{ marginBottom: '2.5rem' }}>
        {messages && messages.length > 0 ? (
          messages.map((msg) => (
            <Message key={msg.id} message={msg} />
          ))
        ) : (
          <div style={{ padding: '1.5rem 0', color: '#6b7280', fontSize: '0.875rem' }}>
            No replies in this thread yet. Be the first to respond.
          </div>
        )}
      </section>

      {/* Reply Section */}
      <section style={{ borderTop: '1px solid #e1e4e8', paddingTop: '1.5rem' }}>
        <ReplyForm onSubmit={handleReply} isSubmitting={isPosting} />
      </section>
    </div>
  );
};
