// Project statuses
export type ProjectStatus =
  | 'CREATED'
  | 'SCRIPT_GENERATING'
  | 'SCRIPT_READY'
  | 'SCENES_GENERATING'
  | 'SCENES_READY'
  | 'VOICE_GENERATING'
  | 'ASSETS_READY'
  | 'SUBTITLES_GENERATING'
  | 'SUBTITLES_READY'
  | 'RENDERING'
  | 'RENDER_READY'
  | 'IN_REVIEW'
  | 'APPROVED'
  | 'REJECTED'
  | 'FAILED'
  | 'CANCELLED';

export const PROCESSING_STATUSES: ProjectStatus[] = [
  'CREATED',
  'SCRIPT_GENERATING',
  'SCENES_GENERATING',
  'VOICE_GENERATING',
  'SUBTITLES_GENERATING',
  'RENDERING',
];

export const STATUS_PIPELINE: ProjectStatus[] = [
  'CREATED',
  'SCRIPT_GENERATING',
  'SCRIPT_READY',
  'SCENES_GENERATING',
  'SCENES_READY',
  'VOICE_GENERATING',
  'ASSETS_READY',
  'SUBTITLES_GENERATING',
  'SUBTITLES_READY',
  'RENDERING',
  'RENDER_READY',
  'IN_REVIEW',
];

export const STATUS_DESCRIPTIONS: Record<ProjectStatus, string> = {
  CREATED: 'Project created, waiting to start',
  SCRIPT_GENERATING: 'AI is generating the script',
  SCRIPT_READY: 'Script complete',
  SCENES_GENERATING: 'Breaking script into scenes',
  SCENES_READY: 'Scenes defined',
  VOICE_GENERATING: 'Generating voice audio',
  ASSETS_READY: 'All assets generated',
  SUBTITLES_GENERATING: 'Creating subtitles',
  SUBTITLES_READY: 'Subtitles complete',
  RENDERING: 'Rendering final video',
  RENDER_READY: 'Video rendered',
  IN_REVIEW: 'Ready for review',
  APPROVED: 'Approved',
  REJECTED: 'Rejected',
  FAILED: 'Failed',
  CANCELLED: 'Cancelled',
};

export type StatusColor = 'blue' | 'yellow' | 'green' | 'red' | 'gray';

export function getStatusColor(status: ProjectStatus): StatusColor {
  switch (status) {
    case 'CREATED':
    case 'SCRIPT_GENERATING':
    case 'SCRIPT_READY':
    case 'SCENES_GENERATING':
    case 'SCENES_READY':
    case 'VOICE_GENERATING':
    case 'ASSETS_READY':
    case 'SUBTITLES_GENERATING':
    case 'SUBTITLES_READY':
    case 'RENDERING':
    case 'RENDER_READY':
      return 'blue';
    case 'IN_REVIEW':
      return 'yellow';
    case 'APPROVED':
      return 'green';
    case 'REJECTED':
    case 'FAILED':
      return 'red';
    case 'CANCELLED':
      return 'gray';
    default:
      return 'gray';
  }
}

export function isProcessing(status: ProjectStatus): boolean {
  return PROCESSING_STATUSES.includes(status);
}

// API Types
export interface Project {
  project_id: string;
  external_id: string;
  status: ProjectStatus;
  current_step?: string;
  topic?: string;
  title?: string;
  created_at: string;
}

export interface Script {
  title: string;
  hook: string;
  setup_text: string;
  build_text: string;
  turning_point_text: string;
  collapse_text: string;
  conclusion_text: string;
  full_script: string;
}

export interface Scene {
  scene_number: number;
  narration_text: string;
  duration_sec: number;
  mood: string;
  keywords: string[];
  visual_prompt: string;
}

export interface AudioFile {
  scene_number: number;
  storage_key: string;
  duration_ms: number;
}

export interface Asset {
  scene_number: number;
  asset_type: string;
  storage_key: string;
  provider: string;
}

export interface Render {
  draft_video_key: string;
  duration_sec: number;
}

export interface Manifest {
  project: Project;
  script: Script | null;
  scenes: Scene[] | null;
  audio_files: AudioFile[] | null;
  assets: Asset[] | null;
  render: Render | null;
}

// Request types
export interface CreateProjectFromTopic {
  topic: string;
  title?: string;
  channel_style: string;
  target_duration_sec: number;
  aspect_ratio: string;
}

export interface CreateProjectFromScript {
  title: string;
  script: string;
  channel_style: string;
  target_duration_sec: number;
  aspect_ratio: string;
}

export type CreateProjectRequest = CreateProjectFromTopic | CreateProjectFromScript;

export interface ReviewRequest {
  notes: string;
  acted_by: string;
}
