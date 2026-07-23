import React from 'react';

export const Footer: React.FC = () => {
  return (
    <footer className="editorial-footer">
      <div className="container footer-inner">
        <div className="footer-links">
          <a href="https://joinchorus.app" target="_blank" rel="noopener noreferrer">
            Website
          </a>
          <span className="footer-dot">&bull;</span>
          <a href="https://docs.joinchorus.app" target="_blank" rel="noopener noreferrer">
            Documentation
          </a>
          <span className="footer-dot">&bull;</span>
          <a href="https://github.com/joinchorus/chorus" target="_blank" rel="noopener noreferrer">
            GitHub
          </a>
          <span className="footer-dot">&bull;</span>
          <a
            href="https://github.com/joinchorus/chorus/blob/main/LICENSE"
            target="_blank"
            rel="noopener noreferrer"
          >
            License
          </a>
          <span className="footer-dot">&bull;</span>
          <span className="footer-version">v0.1.0-alpha</span>
        </div>
      </div>
    </footer>
  );
};
