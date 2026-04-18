interface SkeletonProps {
  width?: string;
  height?: string;
  radius?: string;
  count?: number;
}

export default function Skeleton({ width = '100%', height = '20px', radius = '8px', count = 1 }: SkeletonProps) {
  return (
    <>
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="skeleton"
          style={{ width, height, borderRadius: radius }}
        />
      ))}
    </>
  );
}

export function ProjectCardSkeleton() {
  return (
    <div className="project-card skeleton-card">
      <div className="card-header">
        <Skeleton width="60%" height="22px" />
        <Skeleton width="90px" height="24px" radius="12px" />
      </div>
      <Skeleton width="40%" height="16px" />
      <div className="card-footer" style={{ marginTop: '12px' }}>
        <Skeleton width="100px" height="14px" />
        <Skeleton width="70px" height="14px" />
      </div>
    </div>
  );
}
