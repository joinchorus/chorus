import React from 'react';
import { Link, useLocation } from 'react-router-dom';

export const Navbar: React.FC = () => {
  const location = useLocation();

  return (
    <header
      style={{
        borderBottom: '1px solid #e1e4e8',
        paddingTop: '1.25rem',
        paddingBottom: '1.25rem',
        marginBottom: '2rem',
      }}
    >
      <div
        className="container"
        style={{
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
        }}
      >
        <div>
          <Link
            to="/"
            style={{
              fontSize: '1.25rem',
              fontWeight: 700,
              color: '#111827',
              textDecoration: 'none',
              letterSpacing: '-0.02em',
            }}
          >
            Chorus
          </Link>
          <span
            style={{
              marginLeft: '0.75rem',
              fontSize: '0.875rem',
              color: '#6b7280',
            }}
          >
            Anonymous discussions.
          </span>
        </div>

        <nav style={{ display: 'flex', gap: '1rem', alignItems: 'center' }}>
          <Link
            to="/"
            style={{
              fontSize: '0.875rem',
              color: location.pathname === '/' ? '#111827' : '#0969da',
              fontWeight: location.pathname === '/' ? 600 : 400,
            }}
          >
            Threads
          </Link>
          <Link
            to="/new"
            style={{
              fontSize: '0.875rem',
              color: '#ffffff',
              backgroundColor: '#24292f',
              padding: '5px 12px',
              borderRadius: '4px',
              fontWeight: 500,
              textDecoration: 'none',
            }}
          >
            Create Thread
          </Link>
        </nav>
      </div>
    </header>
  );
};
