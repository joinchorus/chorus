import React, { useState } from 'react';
import { Textarea } from './ui/Textarea';
import { Button } from './ui/Button';

interface ReplyFormProps {
  onSubmit: (content: string, showCountryFlag: boolean) => Promise<void>;
  isSubmitting?: boolean;
}

export const ReplyForm: React.FC<ReplyFormProps> = ({ onSubmit, isSubmitting }) => {
  const [content, setContent] = useState('');
  const [showFlag, setShowFlag] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!content.trim()) {
      setError('Message content cannot be empty.');
      return;
    }
    setError(null);
    try {
      await onSubmit(content, showFlag);
      setContent('');
    } catch (err: any) {
      setError(err.message || 'Failed to post reply.');
    }
  };

  return (
    <form onSubmit={handleSubmit} style={{ marginTop: '2rem' }}>
      <Textarea
        label="Post Reply"
        placeholder="Write your response..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        rows={4}
        error={error || undefined}
      />

      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginTop: '0.75rem',
        }}
      >
        <label
          style={{
            display: 'inline-flex',
            alignItems: 'center',
            gap: '0.5rem',
            fontSize: '0.8125rem',
            color: '#57606a',
            cursor: 'pointer',
          }}
        >
          <input
            type="checkbox"
            checked={showFlag}
            onChange={(e) => setShowFlag(e.target.checked)}
          />
          Display country flag based on IP
        </label>

        <Button type="submit" disabled={isSubmitting || !content.trim()}>
          {isSubmitting ? 'Posting...' : 'Reply'}
        </Button>
      </div>
    </form>
  );
};
