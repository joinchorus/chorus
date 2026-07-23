import React, { useState, useEffect, useCallback } from 'react';
import { fetchNewConversationName } from '../lib/api';

interface OnboardingModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const DEFAULT_NAMES = ['River', 'Echo', 'Ash', 'Stone', 'Willow', 'Cedar', 'Falcon', 'North', 'Quartz', 'Juniper'];

export const OnboardingModal: React.FC<OnboardingModalProps> = ({ isOpen, onClose }) => {
  const [screen, setScreen] = useState<1 | 2 | 3 | 4 | 5 | 6>(1);
  const [conversationName, setConversationName] = useState<string>('');
  const [isGenerating, setIsGenerating] = useState<boolean>(false);
  const [showCountry, setShowCountry] = useState<boolean>(() => {
    return localStorage.getItem('chorus_show_country') === 'true';
  });

  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const isDark = document.documentElement.getAttribute('data-theme') === 'dark' || localStorage.getItem('chorus_theme') === 'dark';
    return isDark ? 'dark' : 'light';
  });

  useEffect(() => {
    if (isOpen) {
      const isDark = document.documentElement.getAttribute('data-theme') === 'dark' || localStorage.getItem('chorus_theme') === 'dark';
      setTheme(isDark ? 'dark' : 'light');
    }
  }, [isOpen]);

  const toggleTheme = () => {
    const nextTheme = theme === 'light' ? 'dark' : 'light';
    setTheme(nextTheme);
    if (nextTheme === 'dark') {
      document.documentElement.setAttribute('data-theme', 'dark');
    } else {
      document.documentElement.removeAttribute('data-theme');
    }
    localStorage.setItem('chorus_theme', nextTheme);
  };

  const generateName = useCallback(async () => {
    setIsGenerating(true);
    try {
      const res = await fetchNewConversationName();
      if (res && res.conversation_name) {
        setConversationName(res.conversation_name);
        localStorage.setItem('chorus_conversation_name', res.conversation_name);
      } else {
        const fallback = DEFAULT_NAMES[Math.floor(Math.random() * DEFAULT_NAMES.length)];
        setConversationName(fallback);
        localStorage.setItem('chorus_conversation_name', fallback);
      }
    } catch {
      const fallback = DEFAULT_NAMES[Math.floor(Math.random() * DEFAULT_NAMES.length)];
      setConversationName(fallback);
      localStorage.setItem('chorus_conversation_name', fallback);
    } finally {
      setIsGenerating(false);
    }
  }, []);

  useEffect(() => {
    if (isOpen) {
      setScreen(1);
      generateName();
    }
  }, [isOpen, generateName]);

  const handleCountryToggle = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.checked;
    setShowCountry(val);
    localStorage.setItem('chorus_show_country', val ? 'true' : 'false');
  };

  const handleFinish = () => {
    localStorage.setItem('chorus_onboarded', 'true');
    localStorage.setItem('chorus_show_country', showCountry ? 'true' : 'false');
    onClose();
  };

  // Keyboard navigation
  useEffect(() => {
    if (!isOpen) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        handleFinish();
      } else if (e.key === 'ArrowRight') {
        if (screen < 6) setScreen((prev) => (prev + 1) as any);
      } else if (e.key === 'ArrowLeft') {
        if (screen > 1) setScreen((prev) => (prev - 1) as any);
      }
    };

    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, [isOpen, screen]);

  if (!isOpen) return null;

  return (
    <div className="onboarding-root" role="dialog" aria-modal="true" aria-labelledby="onboarding-title">
      {/* Editorial Header */}
      <header className="onboarding-editorial-header">
        <div className="onboarding-brand-title">Chorus</div>

        {/* 6 Step Indicators */}
        <div className="onboarding-dots" aria-label={`Slide ${screen} of 6`}>
          {[1, 2, 3, 4, 5, 6].map((stepNum) => (
            <button
              key={stepNum}
              onClick={() => setScreen(stepNum as any)}
              className={`onboarding-dot ${screen === stepNum ? 'onboarding-dot-active' : ''}`}
              aria-label={`Go to slide ${stepNum}`}
            />
          ))}
        </div>

        {/* Theme Switcher & Skip Action */}
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem' }}>
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

          <button onClick={handleFinish} className="onboarding-skip-btn" aria-label="Skip onboarding">
            Skip
          </button>
        </div>
      </header>

      {/* Slide Content Body */}
      <main className="onboarding-editorial-body">
        {/* Slide 1 */}
        {screen === 1 && (
          <div className="onboarding-screen-pane">
            <div className="bubble-illustration-container">
              <svg width="200" height="120" viewBox="0 0 200 120" fill="none">
                <rect x="20" y="20" width="120" height="42" rx="21" fill="var(--bg-subtle)" stroke="var(--border-default)" strokeWidth="1.5" />
                <circle cx="50" cy="41" r="3" fill="var(--text-muted)" />
                <circle cx="68" cy="41" r="3" fill="var(--text-muted)" />
                <circle cx="86" cy="41" r="3" fill="var(--text-muted)" />
                <rect x="70" y="60" width="110" height="42" rx="21" fill="var(--btn-primary-bg)" />
                <line x1="95" y1="81" x2="155" y2="81" stroke="var(--btn-primary-text)" strokeWidth="2.5" strokeLinecap="round" />
              </svg>
            </div>
            <h1 id="onboarding-title" className="editorial-h1">
              Identity belongs to the conversation.
            </h1>
            <p className="editorial-subtitle">
              Not to the person. Conversations matter more than persistent identities.
            </p>
          </div>
        )}

        {/* Slide 2 */}
        {screen === 2 && (
          <div className="onboarding-screen-pane">
            <div className="privacy-illustration-container">
              <svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="var(--text-primary)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2" />
                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
              </svg>
            </div>
            <h1 id="onboarding-title" className="editorial-h1">
              You don't create an account.
            </h1>
            <p className="editorial-subtitle">
              No signups, passwords, emails, or OAuth tracking. You enter, speak, and leave freely.
            </p>
          </div>
        )}

        {/* Slide 3 */}
        {screen === 3 && (
          <div className="onboarding-screen-pane">
            <div className="diagram-illustration-container">
              <div className="diagram-flow">
                <div className="diagram-node node-active">River 🇹🇷</div>
                <div className="diagram-arrow">↓</div>
                <div className="diagram-node node-event">Thread closes</div>
                <div className="diagram-arrow">↓</div>
                <div className="diagram-node node-muted">Name resets</div>
              </div>
            </div>
            <h1 id="onboarding-title" className="editorial-h1">
              Every conversation gives you a temporary name.
            </h1>
            <p className="editorial-subtitle">
              Identity exists only to distinguish participants inside a single thread. The name disappears when you leave.
            </p>
          </div>
        )}

        {/* Slide 4 */}
        {screen === 4 && (
          <div className="onboarding-screen-pane">
            <div className="diagram-illustration-container">
              <div style={{ display: 'flex', gap: '1rem', justifyContent: 'center', margin: '1.5rem 0', opacity: 0.5 }}>
                <span style={{ fontSize: '1.5rem', textDecoration: 'line-through' }}>❤️ Likes</span>
                <span style={{ fontSize: '1.5rem', textDecoration: 'line-through' }}>👥 Followers</span>
                <span style={{ fontSize: '1.5rem', textDecoration: 'line-through' }}>⚡ Ranking</span>
              </div>
            </div>
            <h1 id="onboarding-title" className="editorial-h1">
              Likes don't exist.<br />Followers don't exist.<br />Algorithms don't exist.
            </h1>
            <p className="editorial-subtitle">
              No upvotes, downvotes, reputation scores, or attention engineering. Pure chronological discourse.
            </p>
          </div>
        )}

        {/* Slide 5 */}
        {screen === 5 && (
          <div className="onboarding-screen-pane">
            <div className="privacy-illustration-container">
              <svg width="80" height="80" viewBox="0 0 24 24" fill="none" stroke="var(--text-primary)" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z" />
                <path d="M9 12l2 2 4-4" />
              </svg>
            </div>
            <h1 id="onboarding-title" className="editorial-h1">
              Choose whether to display your country flag.
            </h1>
            <p className="editorial-subtitle" style={{ marginBottom: '1.5rem' }}>
              Show my country only. Never my location.
            </p>
            <label className="editorial-checkbox-label">
              <input
                type="checkbox"
                checked={showCountry}
                onChange={handleCountryToggle}
                className="checkbox-hidden"
              />
              <span className={`checkbox-custom ${showCountry ? 'checkbox-checked' : ''}`}>
                {showCountry && (
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="var(--btn-primary-text)" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round">
                    <polyline points="20 6 9 17 4 12" />
                  </svg>
                )}
              </span>
              <span className="checkbox-label-text">Show my country flag (OFF by default)</span>
            </label>
          </div>
        )}

        {/* Slide 6 */}
        {screen === 6 && (
          <div className="onboarding-screen-pane">
            <h1 id="onboarding-title" className="editorial-h1" style={{ marginBottom: '1.5rem' }}>
              Enter Chorus
            </h1>
            <p className="editorial-subtitle" style={{ marginBottom: '2rem' }}>
              Your temporary conversation identity for this session:
            </p>
            <div className="editorial-card" style={{ marginBottom: '1.5rem' }}>
              <span className="editorial-card-label">Assigned Identity</span>
              <div className="editorial-card-value">
                {isGenerating ? '...' : conversationName || 'River'}
              </div>
            </div>
            <button
              type="button"
              onClick={generateName}
              disabled={isGenerating}
              className="editorial-btn-secondary"
              style={{ marginBottom: '1rem' }}
            >
              {isGenerating ? 'Generating...' : 'Regenerate Name'}
            </button>
          </div>
        )}
      </main>

      {/* Editorial Footer */}
      <footer className="onboarding-editorial-footer">
        {screen > 1 ? (
          <button
            type="button"
            onClick={() => setScreen((prev) => (prev - 1) as any)}
            className="editorial-btn-back"
          >
            ← Back
          </button>
        ) : <div />}

        {screen < 6 ? (
          <button
            type="button"
            onClick={() => setScreen((prev) => (prev + 1) as any)}
            className="editorial-btn-primary"
          >
            Next →
          </button>
        ) : (
          <button
            type="button"
            onClick={handleFinish}
            className="editorial-btn-primary"
          >
            Enter Chorus
          </button>
        )}
      </footer>
    </div>
  );
};
