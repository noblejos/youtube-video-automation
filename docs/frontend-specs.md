# YouTube Video Automation Dashboard - Frontend Specs

Build a minimal, clean web UI for a YouTube video automation API.

**Recommended Stack:** React + TypeScript + Tailwind CSS (or your preferred modern stack)

## API Base URL

```
http://localhost:8080
```

## API Endpoints

### 1. Create Project

```http
POST /projects
Content-Type: application/json
```

**Option A: AI generates script from topic**
```json
{
  "topic": "The Rise of the Zulu Kingdom",
  "title": "Optional video title",
  "channel_style": "dramatic_history_shorts",
  "target_duration_sec": 60,
  "aspect_ratio": "9:16"
}
```

**Option B: User provides full script (skips AI generation)**
```json
{
  "title": "The Great Zimbabwe",
  "script": "Your full script text here...\n\nMultiple paragraphs supported.",
  "channel_style": "dramatic_history_shorts",
  "target_duration_sec": 60,
  "aspect_ratio": "9:16"
}
```

**Response:**
```json
{
  "project_id": "uuid",
  "external_id": "abc123",
  "status": "CREATED",
  "topic": "The Rise of the Zulu Kingdom",
  "title": "Optional video title",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 2. Get Project Details

```http
GET /projects/{project_id}
```

**Response:**
```json
{
  "project_id": "uuid",
  "external_id": "abc123",
  "status": "IN_REVIEW",
  "current_step": "IN_REVIEW",
  "topic": "The Rise of the Zulu Kingdom",
  "title": "Video Title",
  "created_at": "2024-01-01T00:00:00Z"
}
```

### 3. Get Full Project Manifest

```http
GET /projects/{project_id}/manifest
```

**Response:**
```json
{
  "project": {
    "project_id": "uuid",
    "external_id": "abc123",
    "status": "IN_REVIEW",
    "current_step": "IN_REVIEW",
    "topic": "...",
    "title": "...",
    "created_at": "..."
  },
  "script": {
    "title": "Video Title",
    "hook": "What if I told you...",
    "setup_text": "...",
    "build_text": "...",
    "turning_point_text": "...",
    "collapse_text": "...",
    "conclusion_text": "...",
    "full_script": "Complete script text..."
  },
  "scenes": [
    {
      "scene_number": 1,
      "narration_text": "What if I told you that the Zulu Kingdom changed history?",
      "duration_sec": 10.5,
      "mood": "dramatic",
      "keywords": ["zulu", "kingdom", "history"],
      "visual_prompt": "Epic wide shot of African savanna at sunset..."
    }
  ],
  "audio_files": [
    {
      "scene_number": 1,
      "storage_key": "projects/uuid/audio/scene_1.mp3",
      "duration_ms": 10500
    }
  ],
  "assets": [
    {
      "scene_number": 1,
      "asset_type": "image",
      "storage_key": "projects/uuid/images/scene_1.png",
      "provider": "openai"
    }
  ],
  "render": {
    "draft_video_key": "projects/uuid/render/draft.mp4",
    "duration_sec": 60.5
  }
}
```

### 4. Download Video

```http
GET /projects/{project_id}/download
```

**Response:** Binary MP4 file
- `Content-Type: video/mp4`
- `Content-Disposition: attachment; filename="abc123_draft.mp4"`

### 5. Approve Project

```http
POST /projects/{project_id}/approve
Content-Type: application/json
```

**Request:**
```json
{
  "notes": "Looks good!",
  "acted_by": "reviewer@email.com"
}
```

**Response:**
```json
{
  "status": "approved"
}
```

### 6. Reject Project

```http
POST /projects/{project_id}/reject
Content-Type: application/json
```

**Request:**
```json
{
  "notes": "Need changes to scene 3",
  "acted_by": "reviewer@email.com"
}
```

**Response:**
```json
{
  "status": "rejected"
}
```

### 7. Retry Failed Project

```http
POST /projects/{project_id}/retry
```

**Response:**
```json
{
  "status": "retrying"
}
```

### 8. Health Check

```http
GET /health
```

**Response:**
```json
{
  "status": "healthy"
}
```

## Project Status Flow

```
CREATED
    ↓
SCRIPT_GENERATING
    ↓
SCRIPT_READY
    ↓
SCENES_GENERATING
    ↓
SCENES_READY
    ↓
VOICE_GENERATING
    ↓
ASSETS_READY
    ↓
SUBTITLES_GENERATING
    ↓
SUBTITLES_READY
    ↓
RENDERING
    ↓
RENDER_READY
    ↓
IN_REVIEW
    ↓
APPROVED / REJECTED

Special statuses: FAILED, CANCELLED
```

### Status Descriptions

| Status | Description |
|--------|-------------|
| `CREATED` | Project created, waiting to start |
| `SCRIPT_GENERATING` | AI is generating the script |
| `SCRIPT_READY` | Script complete, ready for scenes |
| `SCENES_GENERATING` | Breaking script into scenes |
| `SCENES_READY` | Scenes defined, ready for assets |
| `VOICE_GENERATING` | Generating voice audio for scenes |
| `ASSETS_READY` | All images and audio generated |
| `SUBTITLES_GENERATING` | Creating synchronized subtitles |
| `SUBTITLES_READY` | Subtitles complete |
| `RENDERING` | FFmpeg rendering final video |
| `RENDER_READY` | Video rendered successfully |
| `IN_REVIEW` | Ready for human review |
| `APPROVED` | Approved for publishing |
| `REJECTED` | Rejected, needs changes |
| `FAILED` | Processing failed (can retry) |
| `CANCELLED` | Project cancelled |

## UI Requirements

### 1. Dashboard / Projects List

- List all projects with key info
- Display columns: Title/Topic, Status, Created Date, Progress
- Color-code statuses:
  - **Blue**: Processing (GENERATING, RENDERING)
  - **Yellow**: Needs attention (IN_REVIEW)
  - **Green**: Complete (APPROVED, RENDER_READY)
  - **Red**: Error (FAILED, REJECTED)
- Auto-refresh every 5 seconds to show progress
- Click row to view project details

### 2. Create Project Form

**Mode Toggle:** "Generate from Topic" vs "Custom Script"

**Fields:**
| Field | Type | Required | Notes |
|-------|------|----------|-------|
| Title | text | No (recommended for custom script) | Video title |
| Topic | text | Yes (if no script) | Topic for AI generation |
| Script | textarea | Yes (if custom mode) | Full script text |
| Target Duration | dropdown | No | 30s, 60s, 90s, 120s (default: 60s) |
| Aspect Ratio | dropdown | No | 9:16 Portrait, 16:9 Landscape (default: 9:16) |

**Behavior:**
- Show Topic field when "Generate from Topic" selected
- Show Script textarea when "Custom Script" selected
- Submit button with loading state
- Redirect to project detail on success

### 3. Project Detail View

**Header Section:**
- Title and topic
- Status badge with color
- Created date
- Current step indicator

**Progress Timeline:**
- Visual pipeline showing completed/current/pending steps
- Highlight current step

**Script Section:**
- Collapsible sections for each script part (hook, setup, etc.)
- Or show full script if user-provided

**Scenes Breakdown:**
- Accordion or card list of scenes
- Each scene shows:
  - Scene number
  - Narration text
  - Duration
  - Mood/keywords

**Actions (based on status):**

| Status | Available Actions |
|--------|-------------------|
| `IN_REVIEW` | Video preview, Download, Approve, Reject |
| `FAILED` | View error, Retry button |
| `APPROVED` | Download |

### 4. Video Preview & Download

For `IN_REVIEW` or `RENDER_READY` projects:
- Embedded HTML5 video player
- Source: `/projects/{id}/download`
- Download button (triggers file save)
- Approve/Reject buttons with optional notes textarea

### 5. Error Handling

- Toast notifications for API errors
- Inline error messages on forms
- Retry button for failed projects
- Graceful handling of network issues

## Technical Notes

### CORS

The API may need CORS headers for browser requests. If you encounter CORS issues, the backend will need to be updated to add:

```
Access-Control-Allow-Origin: *
Access-Control-Allow-Methods: GET, POST, OPTIONS
Access-Control-Allow-Headers: Content-Type
```

### Project List Persistence

The API currently doesn't have a `GET /projects` list endpoint. Implement one of these solutions:

**Option A: LocalStorage (Simple)**
- Store created project IDs in localStorage
- Fetch each project individually on dashboard load
- Remove from list if 404

**Option B: Request Backend Endpoint**
- Ask to add `GET /projects` endpoint to backend

### Polling for Updates

```javascript
// Poll project status every 5 seconds
useEffect(() => {
  const interval = setInterval(() => {
    if (isProcessing(project.status)) {
      refetchProject();
    }
  }, 5000);
  return () => clearInterval(interval);
}, [project.status]);

function isProcessing(status) {
  return ['CREATED', 'SCRIPT_GENERATING', 'SCENES_GENERATING', 
          'VOICE_GENERATING', 'RENDERING', 'SUBTITLES_GENERATING']
         .includes(status);
}
```

### Video Download

```javascript
async function downloadVideo(projectId) {
  const response = await fetch(`/projects/${projectId}/download`);
  const blob = await response.blob();
  const url = window.URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = `video_${projectId}.mp4`;
  a.click();
  window.URL.revokeObjectURL(url);
}
```

## Design Guidelines

### Theme
- **Dark mode preferred** (video content looks better)
- Clean, minimal, professional aesthetic
- Consistent spacing and typography

### Colors (Dark Theme Suggestion)
```css
--bg-primary: #0f0f0f;
--bg-secondary: #1a1a1a;
--bg-card: #242424;
--text-primary: #ffffff;
--text-secondary: #a0a0a0;
--accent: #3b82f6;
--success: #22c55e;
--warning: #eab308;
--error: #ef4444;
```

### Responsive
- Mobile-first approach
- Breakpoints: 640px, 768px, 1024px
- Stack cards vertically on mobile
- Collapsible sidebar on tablet

### Loading States
- Skeleton loaders for async data
- Spinner on buttons during submission
- Progress bar for video processing

## Example User Flows

### Flow 1: Create Video from Topic
1. User clicks "New Project"
2. Selects "Generate from Topic"
3. Enters topic: "The Fall of the Roman Empire"
4. Selects 60s duration, 9:16 aspect ratio
5. Clicks "Create"
6. Redirected to project detail
7. Watches progress update through statuses
8. When IN_REVIEW, previews video
9. Clicks "Approve"

### Flow 2: Create Video from Custom Script
1. User clicks "New Project"
2. Selects "Custom Script"
3. Enters title: "Ancient Egypt's Secrets"
4. Pastes full script into textarea
5. Clicks "Create"
6. Project skips SCRIPT_GENERATING, goes to SCENES_GENERATING
7. Continues as normal

### Flow 3: Handle Failed Project
1. User sees project with FAILED status (red)
2. Clicks to view details
3. Sees error message explaining failure
4. Clicks "Retry"
5. Project restarts from failed step
