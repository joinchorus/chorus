import React, { useState, useEffect } from 'react';
import { SYSTEM_BOARDS, fetchNewConversationName } from '../lib/api';

interface CreateThreadFormProps {
  onSubmit: (title: string, boardSlug: string, content: string, showCountryFlag: boolean) => Promise<void>;
  isSubmitting?: boolean;
}

export const CreateThreadForm: React.FC<CreateThreadFormProps> = ({ onSubmit, isSubmitting }) => {
  const [boardSlug, setBoardSlug] = useState('technology');
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [showFlag, setShowFlag] = useState<boolean>(() => {
    return localStorage.getItem('chorus_show_country') === 'true';
  });
  const [tempName, setTempName] = useState<string>('River');
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchNewConversationName().then((res) => {
      if (res && res.conversation_name) {
        setTempName(res.conversation_name);
      }
    }).catch(() => {
      setTempName('River');
    });
  }, []);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setError('Conversation title is required.');
      return;
    }
    setError(null);
    try {
      await onSubmit(title.trim(), boardSlug, content.trim(), showFlag);
    } catch (err: any) {
      setError(err.message || 'Failed to start conversation.');
    }
  };

  return (
    <form onSubmit={handleSubmit} className="create-thread-form">
      {/* Title */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Title
        </label>
        <input
          type="text"
          placeholder="What would you like to discuss?"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          maxLength={120}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '8px', padding: '0.7rem 0.9rem', fontSize: '0.9375rem' }}
        />
      </div>

      {/* Board Selector */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Board
        </label>
        <select
          value={boardSlug}
          onChange={(e) => setBoardSlug(e.target.value)}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '8px', padding: '0.7rem 0.9rem', fontSize: '0.9375rem' }}
        >
          {SYSTEM_BOARDS.map((b) => (
            <option key={b.id} value={b.slug}>
              {b.display_name} — {b.description}
            </option>
          ))}
        </select>
      </div>

      {/* Body */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Body
        </label>
        <textarea
          placeholder="Write your thoughts..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={6}
          maxLength={4000}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '8px', padding: '0.7rem 0.9rem', fontSize: '0.9375rem', fontFamily: 'inherit' }}
        />
      </div>

      {/* Anonymous Identity Note */}
      <div style={{ marginBottom: '1.5rem', padding: '0.85rem 1rem', background: 'var(--bg-subtle)', borderRadius: '8px', border: '1px solid var(--border-subtle)', display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
        <span style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)' }}>
          Posting anonymously as <strong style={{ color: 'var(--text-primary)' }}>{tempName}</strong>
        </span>
        <label style={{ fontSize: '0.8125rem', color: 'var(--text-muted)', display: 'flex', alignItems: 'center', gap: '0.4rem', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={showFlag}
            onChange={(e) => {
              setShowFlag(e.target.checked);
              localStorage.setItem('chorus_show_country', e.target.checked ? 'true' : 'false');
            }}
          />
          Include country flag
        </label>
      </div>

      {error && (
        <div style={{ color: '#ff6b6b', fontSize: '0.875rem', marginBottom: '1rem' }}>
          {error}
        </div>
      )}

      <button
        type="submit"
        disabled={isSubmitting || !title.trim()}
        className="btn-editorial-primary"
        style={{ width: '100%', padding: '0.75rem', fontSize: '0.9375rem' }}
      >
        {isSubmitting ? 'Publishing...' : 'Create Conversation'}
      </button>
    </form>
  );
};
