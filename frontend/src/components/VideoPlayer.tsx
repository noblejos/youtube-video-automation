import { getDownloadUrl } from '../api';

interface VideoPlayerProps {
  projectId: string;
}

export default function VideoPlayer({ projectId }: VideoPlayerProps) {
  const videoUrl = getDownloadUrl(projectId);

  return (
    <div className="video-player" id="video-player">
      <video
        controls
        className="video-element"
        preload="metadata"
      >
        <source src={videoUrl} type="video/mp4" />
        Your browser does not support the video tag.
      </video>
    </div>
  );
}
