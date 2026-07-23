import React from 'react';

export const Footer: React.FC = () => {
  return (
    <footer className="site-footer">
      <div className="container site-footer-inner">
        <div>
          <strong>Chorus</strong> &mdash; Identity belongs to the conversation.
        </div>
        <div style={{ display: 'flex', gap: '1.25rem', alignItems: 'center' }}>
          <a href="https://joinchorus.app" target="_blank" rel="noopener noreferrer" className="text-muted">
            Website
          </a>
          <a href="https://docs.joinchorus.app" target="_blank" rel="noopener noreferrer" className="text-muted">
            Docs
          </a>
          <a href="https://github.com/barissalihbabacan/Chorus" target="_blank" rel="noopener noreferrer" className="text-muted">
            GitHub
          </a>
        </div>
      </div>
    </footer>
  );
};
