import React from 'react';
import type { Message as MessageType } from '../types';
import { Button } from './ui/Button';

interface MessageProps {
  message: MessageType;
}

export const Message: React.FC<MessageProps> = ({ message }) => {
  const formattedDate = new Date(message.created_at).toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  });

  return (
    <article
      style={{
        borderBottom: '1px solid #e1e4e8',
        paddingTop: '1.25rem',
        paddingBottom: '1.25rem',
      }}
    >
      <header
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'baseline',
          marginBottom: '0.625rem',
          fontSize: '0.8125rem',
          color: '#6b7280',
        }}
      >
        <div style={{ display: 'flex', gap: '0.5rem', alignItems: 'center' }}>
          <span className="font-mono" style={{ color: '#24292f', fontWeight: 600 }}>
            {message.author_id}
          </span>
          {message.country && <span title={`Country: ${message.country}`}>[{message.country}]</span>}
          <span>&bull;</span>
          <span>{formattedDate}</span>
        </div>

        {/* Action placeholders */}
        <div style={{ display: 'flex', gap: '0.5rem' }}>
          <Button variant="ghost" size="sm" disabled title="Translation coming soon">
            Translate
          </Button>
          <Button variant="ghost" size="sm" disabled title="Reporting coming soon">
            Report
          </Button>
        </div>
      </header>

      <div
        style={{
          fontSize: '0.9375rem',
          color: '#111827',
          whiteSpace: 'pre-wrap',
          wordBreak: 'break-word',
          lineHeight: '1.6',
        }}
      >
        {message.content}
      </div>
    </article>
  );
};
