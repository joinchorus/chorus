import React, { useState, useEffect } from 'react';
import { Link, useSearchParams } from 'react-router-dom';

interface NavbarProps {
  onOpenOnboarding?: () => void;
}

export const Navbar: React.FC<NavbarProps> = ({ onOpenOnboarding }) => {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    return (localStorage.getItem('chorus_theme') as 'light' | 'dark') || 'dark';
  });
  const [searchParams, setSearchParams] = useSearchParams();
  const searchQuery = searchParams.get('q') || '';

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('chorus_theme', theme);
  }, [theme]);

  const toggleTheme = () => {
    setTheme((prev) => (prev === 'light' ? 'dark' : 'light'));
  };

  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    if (val) {
      setSearchParams({ q: val }, { replace: true });
    } else {
      setSearchParams({}, { replace: true });
    }
  };

  return (
    <header className="editorial-navbar">
      <div className="container navbar-inner">
        {/* Left: Logo & Tagline */}
        <div className="navbar-brand-group">
          <Link to="/" className="navbar-logo">
            Chorus
          </Link>
          <span className="navbar-tagline">Anonymous discussions.</span>
        </div>

        {/* Center: Search */}
        <div className="navbar-search-wrapper">
          <svg className="navbar-search-icon" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <circle cx="11" cy="11" r="8" />
            <line x1="21" y1="21" x2="16.65" y2="16.65" />
          </svg>
          <input
            type="text"
            className="navbar-search-input"
            placeholder="Search discussions..."
            value={searchQuery}
            onChange={handleSearchChange}
          />
        </div>

        {/* Right: Actions & Theme Toggle */}
        <div className="navbar-actions">
          {onOpenOnboarding && (
            <button onClick={onOpenOnboarding} className="navbar-link-btn" title="View Philosophy & Principles">
              Philosophy
            </button>
          )}

          <button
            onClick={toggleTheme}
            className="navbar-icon-btn"
            aria-label={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
            title={theme === 'light' ? 'Switch to dark theme' : 'Switch to light theme'}
          >
            {theme === 'light' ? (
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
              </svg>
            ) : (
              <svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
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

          <Link to="/new" className="navbar-btn-create">
            + Create Conversation
          </Link>
        </div>
      </div>
    </header>
  );
};
