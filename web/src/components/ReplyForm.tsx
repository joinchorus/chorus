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
    <form onSubmit={handleSubmit} className="composer-box" style={{ background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '8px', padding: '1.25rem' }}>
      <Textarea
        placeholder="Write your response to the conversation..."
        value={content}
        onChange={(e) => setContent(e.target.value)}
        rows={4}
        error={error || undefined}
        maxLength={4000}
      />

      <div className="composer-footer" style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', flexWrap: 'wrap', gap: '1rem', marginTop: '1rem' }}>
        <label className="checkbox-label" style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={showFlag}
            onChange={(e) => setShowFlag(e.target.checked)}
            style={{ width: '16px', height: '16px', accentColor: 'var(--accent-blue)' }}
          />
          <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
            Show my country flag <span style={{ color: 'var(--text-muted)', fontSize: '0.75rem' }}>(Show my country only. Never my location.)</span>
          </span>
        </label>

        <Button type="submit" size="md" disabled={isSubmitting || !content.trim()}>
          {isSubmitting ? 'Sending...' : 'Send Message'}
        </Button>
      </div>
    </form>
  );
};
