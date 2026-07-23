import React from 'react';

export interface SkeletonProps {
  className?: string;
  style?: React.CSSProperties;
}

export const Skeleton: React.FC<SkeletonProps> = ({ className = '', style }) => {
  return <div className={`skeleton ${className}`.trim()} style={style} />;
};

export const ThreadSkeleton: React.FC = () => {
  return (
    <div style={{ padding: '1.25rem 0', borderBottom: '1px solid var(--border-subtle)' }}>
      <Skeleton className="skeleton-title" />
      <div style={{ display: 'flex', gap: '0.75rem', alignItems: 'center' }}>
        <Skeleton style={{ width: '80px', height: '14px' }} />
        <Skeleton style={{ width: '120px', height: '14px' }} />
      </div>
    </div>
  );
};
