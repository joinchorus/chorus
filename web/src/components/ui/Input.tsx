import React from 'react';

interface InputProps extends React.InputHTMLAttributes<HTMLInputElement> {
  label?: string;
  error?: string;
}

export const Input: React.FC<InputProps> = ({
  label,
  error,
  id,
  style,
  ...props
}) => {
  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  const inputStyle: React.CSSProperties = {
    width: '100%',
    padding: '8px 12px',
    fontSize: '14px',
    color: '#111827',
    backgroundColor: '#ffffff',
    border: error ? '1px solid #cf222e' : '1px solid #d0d7de',
    borderRadius: '4px',
    outline: 'none',
  };

  return (
    <div style={{ marginBottom: '1rem' }}>
      {label && (
        <label
          htmlFor={inputId}
          style={{
            display: 'block',
            fontSize: '14px',
            fontWeight: 500,
            marginBottom: '6px',
            color: '#111827',
          }}
        >
          {label}
        </label>
      )}
      <input id={inputId} style={{ ...inputStyle, ...style }} {...props} />
      {error && (
        <span style={{ display: 'block', marginTop: '4px', fontSize: '12px', color: '#cf222e' }}>
          {error}
        </span>
      )}
    </div>
  );
};
