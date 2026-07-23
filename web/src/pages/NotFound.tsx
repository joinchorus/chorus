import React from 'react';
import { useNavigate } from 'react-router-dom';
import { EmptyState } from '../components/EmptyState';

export const NotFound: React.FC = () => {
  const navigate = useNavigate();

  return (
    <EmptyState
      title="404 — Page Not Found"
      description="The page you are looking for does not exist or has moved."
      actionLabel="Return to Conversations"
      onAction={() => navigate('/')}
    />
  );
};
