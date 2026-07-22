import React from 'react';

interface TextareaProps extends React.TextareaHTMLAttributes<HTMLTextAreaElement> {
  label?: string;
  error?: string;
}

export const Textarea: React.FC<TextareaProps> = ({
  label,
  error,
  id,
  rows = 4,
  style,
  ...props
}) => {
  const textareaId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined);

  const textareaStyle: React.CSSProperties = {
    width: '100%',
    padding: '8px 12px',
    fontSize: '14px',
    lineHeight: '1.5',
    color: '#111827',
    backgroundColor: '#ffffff',
    border: error ? '1px solid #cf222e' : '1px solid #d0d7de',
    borderRadius: '4px',
    outline: 'none',
    resize: 'vertical',
  };

  return (
    <div style={{ marginBottom: '1rem' }}>
      {label && (
        <label
          htmlFor={textareaId}
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
      <textarea id={textareaId} rows={rows} style={{ ...textareaStyle, ...style }} {...props} />
      {error && (
        <span style={{ display: 'block', marginTop: '4px', fontSize: '12px', color: '#cf222e' }}>
          {error}
        </span>
      )}
    </div>
  );
};
