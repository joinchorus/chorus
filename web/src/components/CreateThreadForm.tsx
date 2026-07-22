import React, { useState } from 'react';
import { Input } from './ui/Input';
import { Textarea } from './ui/Textarea';
import { Button } from './ui/Button';

interface CreateThreadFormProps {
  onSubmit: (title: string, content: string, showCountryFlag: boolean) => Promise<void>;
  isSubmitting?: boolean;
}

export const CreateThreadForm: React.FC<CreateThreadFormProps> = ({ onSubmit, isSubmitting }) => {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [showFlag, setShowFlag] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Thread title is required.');
      return;
    }
    setError(null);
    try {
      await onSubmit(title.trim(), content.trim(), showFlag);
    } catch (err: any) {
      setError(err.message || 'Failed to create thread.');
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <Input
        label="Title"
        placeholder="Discussion title"
        value={title}
        onChange={(e) => setTitle(e.target.value)}
        required
      />

      <Textarea
        label="Body"
        placeholder="Initial message (optional)..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        rows={6}
      />

      {error && (
        <div style={{ marginBottom: '1rem', color: '#cf222e', fontSize: '0.875rem' }}>
          {error}
        </div>
      )}

      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginTop: '1.25rem',
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

        <Button type="submit" disabled={isSubmitting || !title.trim()}>
          {isSubmitting ? 'Creating...' : 'Create Thread'}
        </Button>
      </div>
    </form>
  );
};
