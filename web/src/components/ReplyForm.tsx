import React, { useState } from 'react';
import { Textarea } from './ui/Textarea';
import { Button } from './ui/Button';

interface ReplyFormProps {
  onSubmit: (content: string, showCountryFlag: boolean) => Promise<void>;
  isSubmitting?: boolean;
}

export const ReplyForm: React.FC<ReplyFormProps> = ({ onSubmit, isSubmitting }) => {
  const [content, setContent] = useState('');
  const [showFlag, setShowFlag] = useState<boolean>(() => {
    return localStorage.getItem('chorus_show_country') === 'true';
  });
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
    <form onSubmit={handleSubmit} className="composer-box">
      <Textarea
        placeholder="Write your response to the conversation..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        rows={4}
        error={error || undefined}
        maxLength={4000}
        style={{ width: '100%' }}
      />

      <div className="composer-footer">
        <label className="checkbox-label" style={{ display: 'flex', alignItems: 'center', gap: '0.6rem', cursor: 'pointer', userSelect: 'none' }}>
          <input
            type="checkbox"
            checked={showFlag}
            onChange={(e) => {
              setShowFlag(e.target.checked);
              localStorage.setItem('chorus_show_country', e.target.checked ? 'true' : 'false');
            }}
            style={{ width: '17px', height: '17px', cursor: 'pointer' }}
          />
          <span style={{ fontSize: '0.84375rem', color: 'var(--text-secondary)' }}>
            Show my country flag <span style={{ color: 'var(--text-muted)', fontSize: '0.78125rem' }}>(Show my country only. Never my location.)</span>
          </span>
        </label>

        <Button type="submit" variant="primary" size="md" disabled={isSubmitting || !content.trim()}>
          {isSubmitting ? 'Sending...' : 'Send Message'}
        </Button>
      </div>
    </form>
  );
};
