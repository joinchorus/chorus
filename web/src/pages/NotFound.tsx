import React from 'react';
import { Link } from 'react-router-dom';

export const NotFound: React.FC = () => {
  return (
    <div style={{ padding: '3rem 0', textAlign: 'center' }}>
      <h2 style={{ fontSize: '2rem', fontWeight: 600, color: '#111827', marginBottom: '0.75rem' }}>
        404
      </h2>
      <p style={{ color: '#6b7280', marginBottom: '1.5rem', fontSize: '0.9375rem' }}>
        Page not found.
      </p>
      <Link to="/" style={{ fontSize: '0.875rem', color: '#0969da' }}>
        &larr; Return to Home
      </Link>
    </div>
  );
};
