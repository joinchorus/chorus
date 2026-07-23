import React from 'react';
import type { Thread } from '../types';
import { formatDate } from '../lib/api';
import { OFFICIAL_TOPICS } from '../lib/topics';

interface ThreadHeaderProps {
  thread: Thread;
  participantNames?: string[];
}

export const ThreadHeader: React.FC<ThreadHeaderProps> = ({ thread, participantNames = [] }) => {
  const formattedDate = formatDate(thread.created_at);
  const authorName = thread.conversation_name || 'Anonymous';
  const topicObj = OFFICIAL_TOPICS.find((t) => t.id === thread.topic) || OFFICIAL_TOPICS[0];

  return (
    <header className="thread-detail-header" style={{ padding: '1.75rem 2rem', background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '12px', marginBottom: '2rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', marginBottom: '0.6rem' }}>
        <span style={{ fontSize: '0.75rem', fontWeight: 700, padding: '0.2rem 0.65rem', borderRadius: '4px', background: 'var(--bg-subtle)', border: '1px solid var(--border-default)', color: 'var(--accent-blue)', textTransform: 'uppercase', letterSpacing: '0.04em' }}>
          {topicObj.name}
        </span>
        <span style={{ fontSize: '0.8125rem', color: 'var(--text-muted)' }}>• Started {formattedDate}</span>
      </div>

      <h1 className="thread-detail-title" style={{ fontSize: '1.85rem', fontWeight: 800, letterSpacing: '-0.04em', color: 'var(--text-primary)', margin: '0.4rem 0 1rem', lineHeight: 1.25 }}>
        {thread.title}
      </h1>

      <div style={{ display: 'flex', alignItems: 'center', gap: '1.25rem', flexWrap: 'wrap', fontSize: '0.875rem', color: 'var(--text-secondary)', borderTop: '1px solid var(--border-default)', paddingTop: '1rem', marginTop: '1rem' }}>
        <div>
          Thread Starter: <strong style={{ color: 'var(--text-primary)' }}>{authorName}{thread.country ? ` [${thread.country}]` : ''}</strong>
        </div>

        {participantNames.length > 0 && (
          <div style={{ display: 'flex', alignItems: 'center', gap: '0.4rem' }}>
            <span>Participants:</span>
            {participantNames.map((name, idx) => (
              <span key={idx} style={{ fontSize: '0.75rem', fontWeight: 600, padding: '0.15rem 0.5rem', borderRadius: '12px', background: 'var(--bg-subtle)', border: '1px solid var(--border-default)', color: 'var(--text-primary)' }}>
                {name}
              </span>
            ))}
          </div>
        )}
      </div>
    </header>
  );
};
