import React, { useState, useEffect } from 'react';
import { OFFICIAL_TOPICS } from '../lib/topics';
import { fetchNewConversationName } from '../lib/api';

interface CreateThreadFormProps {
  onSubmit: (title: string, topic: string, content: string, showCountryFlag: boolean) => Promise<void>;
  isSubmitting?: boolean;
}

export const CreateThreadForm: React.FC<CreateThreadFormProps> = ({ onSubmit, isSubmitting }) => {
  const [topic, setTopic] = useState('technology');
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
      await onSubmit(title.trim(), topic, content.trim(), showFlag);
    } catch (err: any) {
      setError(err.message || 'Failed to start conversation.');
    }
  };

  const selectedTopicObj = OFFICIAL_TOPICS.find((t) => t.id === topic) || OFFICIAL_TOPICS[0];

  return (
    <form onSubmit={handleSubmit} className="create-thread-form">
      {/* Topic Dropdown Selector */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Topic Container
        </label>
        <select
          value={topic}
          onChange={(e) => setTopic(e.target.value)}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '6px', padding: '0.6rem 0.8rem', fontSize: '0.9375rem' }}
        >
          {OFFICIAL_TOPICS.map((t) => (
            <option key={t.id} value={t.id}>
              {t.name} — {t.description}
            </option>
          ))}
        </select>
      </div>

      {/* Conversation Title */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Conversation Title
        </label>
        <input
          type="text"
          placeholder="What would you like to discuss?"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          required
          maxLength={120}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '6px', padding: '0.65rem 0.8rem', fontSize: '0.9375rem' }}
        />
      </div>

      {/* Initial Message */}
      <div className="form-group" style={{ marginBottom: '1.5rem' }}>
        <label className="form-label" style={{ display: 'block', fontSize: '0.875rem', fontWeight: 600, color: 'var(--text-secondary)', marginBottom: '0.5rem' }}>
          Opening Message
        </label>
        <textarea
          placeholder="Write your initial post..."
          value={content}
          onChange={(e) => setContent(e.target.value)}
          rows={5}
          maxLength={4000}
          className="input-field"
          style={{ width: '100%', background: 'var(--bg-surface)', color: 'var(--text-primary)', border: '1px solid var(--border-default)', borderRadius: '6px', padding: '0.65rem 0.8rem', fontSize: '0.9375rem', fontFamily: 'inherit' }}
        />
      </div>

      {/* Country Flag Toggle with Explicit Explanation */}
      <div className="form-group" style={{ marginBottom: '1.75rem', padding: '1rem', background: 'var(--bg-surface)', border: '1px solid var(--border-default)', borderRadius: '8px' }}>
        <label style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', cursor: 'pointer' }}>
          <input
            type="checkbox"
            checked={showFlag}
            onChange={(e) => setShowFlag(e.target.checked)}
            style={{ width: '18px', height: '18px', accentColor: 'var(--accent-blue)' }}
          />
          <div>
            <span style={{ fontWeight: 600, color: 'var(--text-primary)', fontSize: '0.875rem' }}>Show Country Flag</span>
            <p style={{ margin: 0, fontSize: '0.75rem', color: 'var(--text-muted)' }}>
              Show my country only. Never my location. (Defaults to OFF for maximum privacy)
            </p>
          </div>
        </label>
      </div>

      {/* Real-time Conversation Live Preview */}
      <div className="preview-card" style={{ marginBottom: '2rem', padding: '1.25rem', background: 'var(--bg-subtle)', border: '1px border-dashed var(--border-default)', borderRadius: '8px' }}>
        <div style={{ fontSize: '0.6875rem', fontWeight: 700, textTransform: 'uppercase', letterSpacing: '0.06em', color: 'var(--text-muted)', marginBottom: '0.75rem' }}>
          Conversation Live Preview
        </div>
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem', marginBottom: '0.35rem' }}>
          <span style={{ fontSize: '0.6875rem', fontWeight: 700, padding: '0.15rem 0.5rem', borderRadius: '4px', background: 'var(--bg-surface)', border: '1px solid var(--border-default)', color: 'var(--accent-blue)' }}>
            {selectedTopicObj.name}
          </span>
          <span style={{ fontSize: '0.75rem', color: 'var(--text-muted)' }}>• Just now</span>
        </div>
        <h4 style={{ fontSize: '1.05rem', fontWeight: 700, color: 'var(--text-primary)', margin: '0.25rem 0' }}>
          {title.trim() || 'Untitled Conversation'}
        </h4>
        <div style={{ fontSize: '0.8125rem', color: 'var(--text-secondary)', marginTop: '0.5rem' }}>
          Participant: <strong style={{ color: 'var(--text-primary)' }}>{tempName}{showFlag ? ' 🇹🇷' : ''}</strong> (Temporary Thread Identity)
        </div>
      </div>

      {error && <div className="form-error" style={{ marginBottom: '1rem', color: 'var(--accent-red)', fontSize: '0.875rem' }}>{error}</div>}

      <div style={{ display: 'flex', justifyContent: 'flex-end', gap: '0.75rem' }}>
        <button
          type="submit"
          disabled={isSubmitting || !title.trim()}
          className="btn btn-primary"
          style={{ background: 'var(--accent-blue)', color: '#0d1117', border: 'none', fontWeight: 700, padding: '0.65rem 1.5rem', borderRadius: '6px', cursor: 'pointer', opacity: isSubmitting || !title.trim() ? 0.5 : 1 }}
        >
          {isSubmitting ? 'Starting...' : 'Start Conversation'}
        </button>
      </div>
    </form>
  );
};
