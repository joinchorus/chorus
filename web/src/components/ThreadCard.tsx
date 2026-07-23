import React from 'react';
import { Link } from 'react-router-dom';
import type { Thread } from '../types';
import { formatDate } from '../lib/api';
import { OFFICIAL_TOPICS } from '../lib/topics';

interface ThreadCardProps {
  thread: Thread;
}

export const ThreadCard: React.FC<ThreadCardProps> = ({ thread }) => {
  const formattedDate = formatDate(thread.created_at);
  const authorName = thread.conversation_name || 'Anonymous';
  const topicObj = OFFICIAL_TOPICS.find((t) => t.id === thread.topic) || OFFICIAL_TOPICS[0];
  const replyCount = thread.message_count !== undefined ? thread.message_count : 1;
  const participantCount = thread.participant_count !== undefined ? thread.participant_count : 1;

  const previewText = thread.body || thread.preview || 'Discussion open for anonymous community input and technical discourse.';

  return (
    <article
      style={{
        padding: '1.75rem 2rem',
        background: 'var(--bg-surface)',
        border: '1px solid var(--border-subtle)',
        borderRadius: '12px',
        marginBottom: '1.5rem',
        transition: 'border-color 0.15s ease, transform 0.15s ease',
      }}
    >
      {/* 1. Topic Badge */}
      <div style={{ marginBottom: '0.75rem' }}>
        <span
          style={{
            fontFamily: 'var(--font-mono)',
            fontSize: '0.75rem',
            fontWeight: 700,
            padding: '0.2rem 0.65rem',
            borderRadius: '4px',
            background: 'var(--bg-subtle)',
            border: '1px solid var(--border-default)',
            color: 'var(--accent-blue)',
            textTransform: 'uppercase',
            letterSpacing: '0.04em',
          }}
        >
          {topicObj.name}
        </span>
      </div>

      {/* 2. Conversation Title */}
      <h3 style={{ fontSize: '1.375rem', fontWeight: 800, margin: '0.35rem 0 0.75rem', lineHeight: 1.3, letterSpacing: '-0.025em' }}>
        <Link to={`/thread/${thread.id}`} style={{ color: 'var(--text-primary)', textDecoration: 'none' }}>
          {thread.title}
        </Link>
      </h3>

      {/* 3. Three-line message preview */}
      <p
        style={{
          fontSize: '1rem',
          color: 'var(--text-secondary)',
          lineHeight: 1.65,
          marginBottom: '1.25rem',
          display: '-webkit-box',
          WebkitLineClamp: 3,
          WebkitBoxOrient: 'vertical',
          overflow: 'hidden',
          textOverflow: 'ellipsis',
        }}
      >
        {previewText}
      </p>

      {/* 4. Footer Metadata */}
      <div
        style={{
          fontSize: '0.84375rem',
          color: 'var(--text-muted)',
          display: 'flex',
          alignItems: 'center',
          flexWrap: 'wrap',
          gap: '0.6rem',
        }}
      >
        <span style={{ color: 'var(--text-secondary)', fontWeight: 500 }}>
          Started by <strong style={{ color: 'var(--text-primary)', fontWeight: 600 }}>{authorName}{thread.country ? ` [${thread.country}]` : ''}</strong>
        </span>
        <span>&bull;</span>
        <span>{replyCount} {replyCount === 1 ? 'reply' : 'replies'}</span>
        <span>&bull;</span>
        <span>{participantCount} {participantCount === 1 ? 'participant' : 'participants'}</span>
        <span>&bull;</span>
        <span>{formattedDate}</span>
      </div>
    </article>
  );
};
