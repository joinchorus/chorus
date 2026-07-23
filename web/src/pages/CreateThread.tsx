import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createThread } from '../lib/api';
import { CreateThreadForm } from '../components/CreateThreadForm';

export const CreateThread: React.FC = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCreate = async (title: string, topic: string, body: string, showCountry: boolean) => {
    setIsSubmitting(true);
    try {
      const storedName = localStorage.getItem('chorus_conversation_name') || undefined;
      const thread = await createThread({
        title,
        topic,
        body,
        show_country: showCountry,
        conversation_name: storedName,
      });

      navigate(`/thread/${thread.id}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div style={{ width: '100%', padding: '1rem 0' }}>
      <div style={{ marginBottom: '2rem' }}>
        <button
          type="button"
          onClick={() => navigate('/')}
          style={{ background: 'none', border: 'none', color: 'var(--text-secondary)', cursor: 'pointer', fontSize: '0.875rem', fontWeight: 600, padding: 0, marginBottom: '1rem' }}
        >
          ← Back to Conversations
        </button>
        <h2 style={{ fontSize: '1.75rem', fontWeight: 800, letterSpacing: '-0.035em', color: 'var(--text-primary)' }}>
          Start a Conversation
        </h2>
        <p style={{ color: 'var(--text-secondary)', fontSize: '0.9375rem', marginTop: '0.25rem' }}>
          Identity belongs to the conversation. You will post under a temporary thread identity.
        </p>
      </div>

      <CreateThreadForm onSubmit={handleCreate} isSubmitting={isSubmitting} />
    </div>
  );
};
