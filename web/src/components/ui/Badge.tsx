import React from 'react';

export interface BadgeProps {
  children: React.ReactNode;
  variant?: 'default' | 'name';
  className?: string;
}

export const Badge: React.FC<BadgeProps> = ({
  children,
  variant = 'default',
  className = '',
}) => {
  const variantClass = variant === 'name' ? 'badge-name' : '';
  return (
    <span className={`badge ${variantClass} ${className}`.trim()}>
      {children}
    </span>
  );
};
