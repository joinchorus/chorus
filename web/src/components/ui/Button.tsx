import React from 'react';

interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: 'primary' | 'secondary' | 'ghost';
  size?: 'sm' | 'md';
}

export const Button: React.FC<ButtonProps> = ({
  children,
  variant = 'primary',
  size = 'md',
  disabled,
  className = '',
  style,
  ...props
}) => {
  const baseStyle: React.CSSProperties = {
    display: 'inline-flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontWeight: 500,
    cursor: disabled ? 'not-allowed' : 'pointer',
    borderRadius: '4px',
    border: '1px solid transparent',
    transition: 'background-color 0.15s ease, border-color 0.15s ease',
    opacity: disabled ? 0.6 : 1,
    fontSize: size === 'sm' ? '13px' : '14px',
    padding: size === 'sm' ? '4px 10px' : '6px 14px',
  };

  let variantStyle: React.CSSProperties = {};
  if (variant === 'primary') {
    variantStyle = {
      backgroundColor: '#24292f',
      color: '#ffffff',
      borderColor: '#24292f',
    };
  } else if (variant === 'secondary') {
    variantStyle = {
      backgroundColor: '#f6f8fa',
      color: '#24292f',
      borderColor: '#d0d7de',
    };
  } else if (variant === 'ghost') {
    variantStyle = {
      backgroundColor: 'transparent',
      color: '#57606a',
      borderColor: 'transparent',
    };
  }

  return (
    <button
      disabled={disabled}
      style={{ ...baseStyle, ...variantStyle, ...style }}
      className={className}
      {...props}
    >
      {children}
    </button>
  );
};
