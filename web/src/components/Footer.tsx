import React from 'react';

export const Footer: React.FC = () => {
  return (
    <footer
      style={{
        borderTop: '1px solid #e1e4e8',
        marginTop: '4rem',
        paddingTop: '1.5rem',
        paddingBottom: '2rem',
        fontSize: '0.8125rem',
        color: '#6b7280',
      }}
    >
      <div className="container" style={{ display: 'flex', justifyContent: 'space-between' }}>
        <div>Chorus &mdash; Privacy first. Anonymous by default.</div>
        <div>Separation of Concerns &amp; SSOT Architecture</div>
      </div>
    </footer>
  );
};
