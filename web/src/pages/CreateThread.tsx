import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { createThread } from '../lib/api';
import { CreateThreadForm } from '../components/CreateThreadForm';

export const CreateThread: React.FC = () => {
  const navigate = useNavigate();
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleCreate = async (title: string, body: string, showCountry: boolean) => {
    setIsSubmitting(true);
    try {
      const thread = await createThread({
        title,
        body,
        show_country: showCountry,
      });

      navigate(`/thread/${thread.id}`);
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div style={{ maxWidth: '680px' }}>
      <h2
        style={{
          fontSize: '1.25rem',
          fontWeight: 600,
          color: '#111827',
          marginBottom: '1.5rem',
          borderBottom: '1px solid #e1e4e8',
          paddingBottom: '0.75rem',
        }}
      >
        Create Thread
      </h2>

      <CreateThreadForm onSubmit={handleCreate} isSubmitting={isSubmitting} />
    </div>
  );
};
