import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';

export const Navbar: React.FC = () => {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('chorus_theme') as 'light' | 'dark') || 'light';
  });

  useEffect(() => {
    if (theme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
    localStorage.setItem('chorus_theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme((prev) => (prev === 'light' ? 'dark' : 'light'));
  };

  return (
    <header className="site-header">
      <div className="container site-nav-container">
        <div className="site-logo-group">
          <Link to="/" className="site-logo">
            Chorus
          </Link>
          <span className="site-tagline">
            Anonymous discussions.
          </span>
        </div>

        <nav className="site-nav-links" aria-label="Main Navigation" style={{ display: 'flex', alignItems: 'center', gap: '1.25rem' }}>
          <a
            href="https://joinchorus.app"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '0.875rem', fontWeight: 600 }}
          >
            Website
          </a>

          <a
            href="https://docs.joinchorus.app"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '0.875rem', fontWeight: 600 }}
          >
            Documentation
          </a>

          <a
            href="https://github.com/chorus-project/chorus"
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: 'var(--text-secondary)', textDecoration: 'none', fontSize: '0.875rem', fontWeight: 600 }}
          >
            GitHub
          </a>

          <button
            onClick={toggleTheme}
            className="theme-toggle-btn"
            aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
            title={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
          >
            {theme === 'light' ? (
              /* Moon Icon for Light Mode */
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              </svg>
            ) : (
              /* Sun Icon for Dark Mode */
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <circle cx="12" cy="12" r="5" />
                <line x1="12" y1="1" x2="12" y2="3" />
                <line x1="12" y1="21" x2="12" y2="23" />
                <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
                <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
                <line x1="1" y1="12" x2="3" y2="12" />
                <line x1="21" y1="12" x2="23" y2="12" />
                <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
                <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
              </svg>
            )}
          </button>
        </nav>
      </div>
    </header>
  );
};
