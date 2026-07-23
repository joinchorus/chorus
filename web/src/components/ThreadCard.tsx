import React from 'react';
import { Link } from 'react-router-dom';
import type { Thread } from '../types';
import { formatDate } from '../lib/api';

interface ThreadCardProps {
  thread: Thread;
}

export const ThreadCard: React.FC<ThreadCardProps> = ({ thread }) => {
  const formattedDate = formatDate(thread.created_at);
  const authorName = thread.conversation_name || 'Anonymous';
  const replyCount = thread.message_count !== undefined ? thread.message_count : 0;
  const participantCount = thread.participant_count !== undefined ? thread.participant_count : 1;
  const boardSlug = thread.board_slug || (thread.topic || 'technology').toLowerCase();
  const boardName = (thread.board_display_name || thread.topic || 'Technology').toUpperCase();
  const countryTag = thread.country ? `[${thread.country}]` : '';

  const excerpt =
    thread.body ||
    thread.preview ||
    'Discussion open for anonymous community input and technical discourse.';

  return (
    <article className="editorial-card">
      <div className="card-topic">
        <Link to={`/board/${boardSlug}`} className="card-topic-link">
          {boardName}
        </Link>
      </div>
      <h2 className="card-title">
        <Link to={`/thread/${thread.id}`}>{thread.title}</Link>
      </h2>
      <p className="card-excerpt">{excerpt}</p>

      <div className="card-meta">
        <span className="card-meta-author">
          Started by {authorName} {countryTag}
        </span>
        <span className="card-meta-dot">&bull;</span>
        <span className="card-meta-stat">
          {replyCount} {replyCount === 1 ? 'reply' : 'replies'}
        </span>
        <span className="card-meta-dot">&bull;</span>
        <span className="card-meta-stat">
          {participantCount} {participantCount === 1 ? 'participant' : 'participants'}
        </span>
        <span className="card-meta-dot">&bull;</span>
        <span className="card-meta-time">{formattedDate}</span>
      </div>
    </article>
  );
};
