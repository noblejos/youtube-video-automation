# YouTube Video Automation System - Complete Specifications

## 1. System Overview

### 1.1 Purpose
Automated generation of faceless YouTube history videos from either AI-generated or user-provided scripts. The system orchestrates a multi-stage pipeline: script generation → scene breakdown → voice synthesis → image generation → subtitle creation → video rendering → human review.

### 1.2 Target Use Case
Content creators producing short-form (30-120 second) vertical video content for YouTube Shorts, TikTok, or Instagram Reels focused on historical topics.

### 1.3 Key Features
- **Dual Input Modes**: Generate script from topic OR use custom script
- **AI-Powered Generation**: LLM-based script and scene generation
- **Multi-Provider Support**: Swappable voice (Polly/mock), image (DALL-E/mock), LLM (GPT-4/mock)
- **Parallel Processing**: Voice and image generation happen simultaneously per scene
- **Automatic Retries**: Failed jobs retry up to 3 times
- **Human Review Gate**: Projects require approval before publishing
- **Ken Burns Effect**: Cinematic zoom/pan on static images
- **Subtitle Generation**: Automatic SRT subtitles from audio timing

---

## 2. Data Models

### 2.1 Project

Represents a video generation project.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `external_id` | String | Yes | Short 8-char ID for external reference (e.g., "abc123") |
| `topic` | String | Yes* | Topic for AI generation (*if script not provided) |
| `title` | String | No | Video title (optional, defaults to topic) |
| `channel_style` | String | Yes | Content style (default: "dramatic_history_shorts") |
| `target_duration_sec` | Integer | Yes | Target video length in seconds (default: 120) |
| `aspect_ratio` | String | Yes | Video aspect ratio (default: "9:16") |
| `status` | String | Yes | Current project status (see 2.1.1) |
| `review_required` | Boolean | Yes | Whether human review is required (default: true) |
| `current_step` | String | No | Human-readable current step (e.g., "SCRIPT_GENERATION") |
| `error_message` | String | No | Error details if project failed |
| `created_at` | Timestamp | Yes | Creation timestamp |
| `updated_at` | Timestamp | Yes | Last update timestamp |

#### 2.1.1 Project Status Values

| Status | Description |
|--------|-------------|
| `CREATED` | Project created, queued for script generation |
| `SCRIPT_GENERATING` | LLM generating script from topic |
| `SCRIPT_READY` | Script complete, ready for scene breakdown |
| `SCENES_GENERATING` | LLM breaking script into scenes |
| `SCENES_READY` | Scenes defined, ready for asset generation |
| `VOICE_GENERATING` | TTS generating voice-overs |
| `VOICE_READY` | All voice files generated (legacy, same as ASSETS_GENERATING) |
| `ASSETS_GENERATING` | Images being generated |
| `ASSETS_READY` | All voice and image assets complete |
| `SUBTITLES_GENERATING` | Creating SRT subtitles |
| `SUBTITLES_READY` | Subtitles complete |
| `RENDERING` | FFmpeg rendering final video |
| `RENDER_READY` | Video rendered, ready for review |
| `REVIEW_PACKAGED` | Review package prepared |
| `IN_REVIEW` | Awaiting human approval/rejection |
| `APPROVED` | Approved for publishing |
| `REJECTED` | Rejected by reviewer |
| `PUBLISHING` | Publishing to YouTube (future) |
| `PUBLISHED` | Published successfully (future) |
| `FAILED` | Processing failed permanently |
| `CANCELLED` | Project cancelled by user (future) |

#### 2.1.2 Channel Style Options

| Value | Description |
|-------|-------------|
| `dramatic_history_shorts` | Default style: dramatic narration for history content |
| `educational_explainer` | Educational, informative tone |
| `mystery_thriller` | Suspenseful, mysterious narration |

#### 2.1.3 Aspect Ratio Options

| Value | Dimensions | Use Case |
|-------|------------|----------|
| `9:16` | 1080x1920 | YouTube Shorts, TikTok, Instagram Reels (default) |
| `16:9` | 1920x1080 | YouTube standard, landscape |
| `1:1` | 1080x1080 | Instagram square posts |

### 2.2 Script

Represents a generated or user-provided script with structured narrative sections.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `hook` | String | No | Opening hook (3-5 seconds) |
| `setup_text` | String | No | Setup/context section |
| `build_text` | String | No | Building tension section |
| `turning_point_text` | String | No | Climax/turning point |
| `collapse_text` | String | No | Resolution/collapse section |
| `conclusion_text` | String | No | Closing/conclusion |
| `full_script` | String | Yes | Complete concatenated script text |
| `raw_model_response` | JSONB | No | Raw LLM response for debugging |
| `created_at` | Timestamp | Yes | Creation timestamp |
| `updated_at` | Timestamp | Yes | Last update timestamp |

**Notes:**
- For user-provided scripts, `full_script` is populated and section fields may be empty
- For AI-generated scripts, LLM returns structured sections that are concatenated into `full_script`
- Section fields follow narrative structure for dramatic history content

### 2.3 Scene

Represents a single scene in the video with narration, timing, and visual instructions.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `scene_number` | Integer | Yes | Scene order (1-indexed) |
| `status` | String | Yes | PENDING, READY, FAILED |
| `story_role` | String | No | Narrative role: hook, setup, build, turning_point, collapse, conclusion |
| `energy_level` | String | No | Pacing: low, medium, high |
| `narration_text` | String | Yes | Exact text for voice-over |
| `ssml_text` | String | No | SSML-formatted narration (future) |
| `duration_sec` | Numeric(6,2) | Yes | Scene duration in seconds |
| `start_time_sec` | Numeric(6,2) | No | Start time in final video |
| `mood` | String | Yes | Emotional tone (e.g., "dramatic", "mysterious", "triumphant") |
| `keywords` | JSONB | No | JSON array of visual keywords |
| `visual_prompt` | String | Yes | Detailed DALL-E prompt for image generation |
| `negative_prompt` | String | No | Elements to avoid in image generation |
| `camera_motion` | String | No | Ken Burns effect: slow_zoom_in, pan_left, pan_right, zoom_out |
| `transition_in` | String | No | Entry transition: fade, cut, dissolve |
| `transition_out` | String | No | Exit transition: fade, cut, dissolve |
| `asset_strategy` | String | Yes | Asset sourcing: ai_or_archive, ai_only (default: ai_or_archive) |
| `transition_type` | String | No | Legacy transition field |
| `audio_config` | JSONB | No | Audio configuration (voice, background music) |
| `created_at` | Timestamp | Yes | Creation timestamp |
| `updated_at` | Timestamp | Yes | Last update timestamp |

**Constraints:**
- `(project_id, scene_number)` must be unique
- Typical scene count: 6-8 scenes for 120-second videos
- Scene duration calculated based on narration text length (avg 150 words/minute)

### 2.4 AudioFile

Represents a generated voice-over file for a scene.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `scene_id` | UUID | Yes | Foreign key to scene |
| `provider` | String | Yes | TTS provider: mock, polly |
| `voice_id` | String | Yes | Voice identifier (e.g., "Matthew", "Ayanda") |
| `engine` | String | No | TTS engine: standard, neural |
| `storage_key` | String | Yes | Relative path to audio file (e.g., "projects/{id}/audio/scene_001.mp3") |
| `duration_ms` | Integer | No | Audio duration in milliseconds |
| `speech_marks_key` | String | No | Path to speech marks JSON (for word-level timing) |
| `created_at` | Timestamp | Yes | Creation timestamp |

**Notes:**
- One audio file per scene
- MP3 format, mono, 22050 Hz
- Speech marks contain word-level timing for subtitle synchronization

### 2.5 Asset

Represents a generated media asset (image, video, audio).

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `scene_id` | UUID | No | Foreign key to scene (null for project-level assets) |
| `asset_type` | String | Yes | IMAGE, AUDIO, VIDEO, SUBTITLE, BACKGROUND |
| `provider` | String | Yes | Generation provider: mock, openai, dalle, etc. |
| `storage_key` | String | Yes | Relative path to asset file |
| `mime_type` | String | No | MIME type (e.g., "image/png") |
| `source_url` | String | No | Original source URL if from archive |
| `prompt_used` | String | No | Generation prompt used |
| `metadata` | JSONB | No | Provider-specific metadata |
| `created_at` | Timestamp | Yes | Creation timestamp |

**Notes:**
- Images are PNG, 1080x1920 for 9:16 aspect ratio
- One image asset per scene

### 2.6 Subtitle

Represents generated subtitle files.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `format` | String | Yes | Subtitle format: srt, vtt, ass (default: srt) |
| `storage_key` | String | Yes | Relative path to subtitle file |
| `created_at` | Timestamp | Yes | Creation timestamp |

**SRT Format Example:**
```srt
1
00:00:00,000 --> 00:00:03,500
What if I told you that the Zulu Kingdom

2
00:00:03,500 --> 00:00:07,200
changed the course of African history?
```

### 2.7 Render

Represents a rendered video output.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `render_type` | String | Yes | draft, final, preview |
| `storage_key` | String | Yes | Relative path to video file |
| `metadata` | JSONB | No | Render metadata (duration, resolution, etc.) |
| `created_at` | Timestamp | Yes | Creation timestamp |

**Metadata Example:**
```json
{
  "duration_sec": 62.5,
  "resolution": "1080x1920",
  "fps": 30,
  "codec": "h264",
  "bitrate": "2000k"
}
```

### 2.8 Job

Represents a background job in the processing queue.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `job_type` | String | Yes | Job type constant (see 2.8.1) |
| `status` | String | Yes | QUEUED, RUNNING, SUCCEEDED, FAILED, RETRYING, CANCELLED |
| `attempt_count` | Integer | Yes | Number of execution attempts (default: 0) |
| `max_attempts` | Integer | Yes | Maximum retry attempts (default: 3) |
| `payload` | JSONB | Yes | Job-specific input data |
| `result` | JSONB | No | Job execution result |
| `error_message` | String | No | Error details if failed |
| `created_at` | Timestamp | Yes | Creation timestamp |
| `updated_at` | Timestamp | Yes | Last update timestamp |

#### 2.8.1 Job Types

| Job Type | Description | Payload Fields |
|----------|-------------|----------------|
| `SCRIPT_GENERATION` | Generate script from topic | `topic`, `channel_style`, `target_duration` |
| `SCENES_GENERATION` | Break script into scenes | `script_id` |
| `VOICE_GENERATION` | Generate voice-over for scene | `scene_id`, `voice_id`, `engine` |
| `IMAGE_GENERATION` | Generate image for scene | `scene_id`, `aspect_ratio`, `style_profile` |
| `SUBTITLE_GENERATION` | Create SRT subtitles | - |
| `RENDER` | Render final video | - |
| `REVIEW_PACKAGE` | Package assets for review | - |

### 2.9 ReviewAction

Represents a human review action on a project.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | UUID | Yes | Primary key |
| `project_id` | UUID | Yes | Foreign key to project |
| `action` | String | Yes | APPROVE, REJECT, RERENDER |
| `notes` | String | No | Reviewer notes |
| `acted_by` | String | No | Reviewer identifier (email) |
| `created_at` | Timestamp | Yes | Action timestamp |

---

## 3. API Specifications

### 3.1 Create Project

**Endpoint:** `POST /projects`

**Description:** Creates a new video generation project. Supports two modes: AI script generation from topic, or user-provided custom script.

**Request Body:**
```json
{
  "topic": "The Rise and Fall of Mansa Musa",
  "title": "Mansa Musa: The Richest Man in History",
  "script": "Optional: Full script text here...",
  "channel_style": "dramatic_history_shorts",
  "target_duration_sec": 120,
  "aspect_ratio": "9:16",
  "review_required": true
}
```

**Request Fields:**

| Field | Type | Required | Default | Validation |
|-------|------|----------|---------|------------|
| `topic` | string | Yes* | - | Required if `script` not provided |
| `title` | string | No | topic value | Max 200 chars |
| `script` | string | No | - | If provided, skips SCRIPT_GENERATION |
| `channel_style` | string | No | "dramatic_history_shorts" | - |
| `target_duration_sec` | integer | No | 120 | 30-300 seconds |
| `aspect_ratio` | string | No | "9:16" | "9:16", "16:9", "1:1" |
| `review_required` | boolean | No | true | - |

**Response:** `201 Created`
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "external_id": "a1b2c3d4",
  "status": "CREATED",
  "topic": "The Rise and Fall of Mansa Musa",
  "title": "Mansa Musa: The Richest Man in History",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `400 Bad Request`: Missing required fields or validation error
- `500 Internal Server Error`: Server error

### 3.2 Get Project Status

**Endpoint:** `GET /projects/{project_id}`

**Description:** Returns current project status and metadata.

**Response:** `200 OK`
```json
{
  "project_id": "550e8400-e29b-41d4-a716-446655440000",
  "external_id": "a1b2c3d4",
  "status": "SCENES_GENERATING",
  "current_step": "SCENES_GENERATION",
  "topic": "The Rise and Fall of Mansa Musa",
  "title": "Mansa Musa: The Richest Man in History",
  "created_at": "2024-01-15T10:30:00Z"
}
```

**Error Responses:**
- `404 Not Found`: Project not found

### 3.3 Get Project Manifest

**Endpoint:** `GET /projects/{project_id}/manifest`

**Description:** Returns complete project data including script, scenes, assets, and render info.

**Response:** `200 OK`
```json
{
  "project": {
    "project_id": "550e8400-e29b-41d4-a716-446655440000",
    "external_id": "a1b2c3d4",
    "status": "IN_REVIEW",
    "current_step": "IN_REVIEW",
    "topic": "The Rise and Fall of Mansa Musa",
    "title": "Mansa Musa: The Richest Man in History",
    "created_at": "2024-01-15T10:30:00Z"
  },
  "script": {
    "hook": "What if I told you about a king so wealthy...",
    "setup_text": "In the 14th century, Mansa Musa ruled...",
    "build_text": "His legendary pilgrimage to Mecca...",
    "turning_point_text": "But his generosity had unexpected consequences...",
    "collapse_text": "The empire began to fracture...",
    "conclusion_text": "Today, his legacy lives on...",
    "full_script": "Complete concatenated script..."
  },
  "scenes": [
    {
      "scene_number": 1,
      "narration_text": "What if I told you about a king so wealthy...",
      "duration_sec": 8.5,
      "mood": "mysterious",
      "keywords": ["king", "wealth", "africa", "gold"],
      "visual_prompt": "Cinematic wide shot of West African savanna at golden hour...",
      "asset_strategy": "ai_or_archive"
    }
  ],
  "audio_files": [
    {
      "scene_number": 1,
      "storage_key": "projects/550e8400.../audio/scene_001.mp3",
      "duration_ms": 8500
    }
  ],
  "assets": [
    {
      "scene_number": 1,
      "asset_type": "IMAGE",
      "storage_key": "projects/550e8400.../images/scene_001.png",
      "provider": "openai"
    }
  ],
  "render": {
    "draft_video_key": "projects/550e8400.../renders/draft.mp4",
    "duration_sec": 62.5
  }
}
```

### 3.4 Download Video

**Endpoint:** `GET /projects/{project_id}/download`

**Description:** Downloads the rendered video file.

**Response:** `200 OK`
- Content-Type: `video/mp4`
- Content-Disposition: `attachment; filename="a1b2c3d4_draft.mp4"`
- Body: Binary MP4 file

**Error Responses:**
- `404 Not Found`: Project or video not found
- `400 Bad Request`: Video not yet rendered

### 3.5 Approve Project

**Endpoint:** `POST /projects/{project_id}/approve`

**Description:** Approves a project for publishing.

**Request Body:**
```json
{
  "notes": "Looks great! Ready to publish.",
  "acted_by": "reviewer@company.com"
}
```

**Response:** `200 OK`
```json
{
  "status": "approved"
}
```

**Error Responses:**
- `400 Bad Request`: Project not in reviewable state
- `404 Not Found`: Project not found

### 3.6 Reject Project

**Endpoint:** `POST /projects/{project_id}/reject`

**Description:** Rejects a project with feedback.

**Request Body:**
```json
{
  "notes": "Scene 3 image doesn't match narration. Please regenerate.",
  "acted_by": "reviewer@company.com"
}
```

**Response:** `200 OK`
```json
{
  "status": "rejected"
}
```

### 3.7 Retry Failed Project

**Endpoint:** `POST /projects/{project_id}/retry`

**Description:** Re-queues all failed jobs for a project.

**Response:** `200 OK`
```json
{
  "status": "retrying",
  "jobs_retried": 2
}
```

**Error Responses:**
- `400 Bad Request`: No failed jobs found
- `404 Not Found`: Project not found

### 3.8 Health Check

**Endpoint:** `GET /health`

**Response:** `200 OK`
```json
{
  "status": "healthy"
}
```

---

## 4. Processing Rules

### 4.1 Script Generation Rules

**Input:** Topic string (e.g., "The Rise and Fall of Mansa Musa")

**LLM Prompt Structure:**
```
Generate a dramatic short-form script about: {topic}

Target duration: {target_duration_sec} seconds
Style: {channel_style}

Structure the script into 6 sections:
1. Hook (3-5 seconds): Attention-grabbing opening
2. Setup (15-20%): Context and background
3. Build (30-35%): Rising action
4. Turning Point (20-25%): Climax
5. Collapse (15-20%): Resolution
6. Conclusion (5-10%): Closing message

Output as JSON with fields: hook, setup_text, build_text, turning_point_text, collapse_text, conclusion_text
```

**Output Processing:**
- Parse LLM JSON response
- Concatenate sections into `full_script`
- Save to `scripts` table
- Extract title from response or use topic

### 4.2 Scene Generation Rules

**Input:** Script text, target duration

**Scene Count Calculation:**
```
scene_count = ceil(target_duration_sec / avg_scene_duration)
avg_scene_duration = 8-12 seconds
typical_scene_count = 6-8 for 120-second videos
```

**LLM Prompt Structure:**
```
Break this script into {scene_count} scenes for a {target_duration_sec}-second video:

{full_script}

For each scene, provide:
- scene_number: Sequential number
- story_role: hook, setup, build, turning_point, collapse, or conclusion
- energy_level: low, medium, or high
- narration_text: Exact words to be spoken
- duration_sec: Estimated duration
- mood: Emotional tone (dramatic, mysterious, triumphant, etc.)
- keywords: Array of 3-5 visual keywords
- visual_prompt: Detailed DALL-E prompt for image generation (describe setting, lighting, composition, style)
- camera_motion: slow_zoom_in, pan_left, pan_right, or zoom_out

Output as JSON array.
```

**Scene Validation Rules:**
- Total scene duration should match target_duration_sec ±10%
- Each scene must have narration_text
- visual_prompt must be detailed (50+ characters)
- Scene numbers must be sequential starting at 1

### 4.3 Voice Generation Rules

**Per Scene:**
- Input: `narration_text`, `voice_id`, `engine`
- Provider: AWS Polly or mock
- Output format: MP3, mono, 22050 Hz
- Speech marks: Optional JSON with word-level timing

**Voice Selection:**
- Default: Configurable via `POLLY_DEFAULT_VOICE` (e.g., "Ayanda", "Matthew")
- Engine: "neural" (higher quality) or "standard"

**Duration Calculation:**
```
words_per_minute = 150 (average speaking pace)
duration_sec = (word_count / words_per_minute) * 60
```

### 4.4 Image Generation Rules

**Per Scene:**
- Input: `visual_prompt`, `aspect_ratio`, `style_profile`
- Provider: DALL-E 3 or mock
- Output format: PNG
- Resolution: 1080x1920 (9:16), 1920x1080 (16:9), or 1080x1080 (1:1)

**Prompt Enhancement:**
```
base_prompt = scene.visual_prompt
enhanced_prompt = f"{base_prompt}, cinematic lighting, high detail, dramatic composition, professional photography"

if channel_style == "dramatic_history_shorts":
    enhanced_prompt += ", historical accuracy, period-appropriate"
```

**Negative Prompt:**
```
"text, watermarks, logos, modern objects, anachronisms, low quality"
```

### 4.5 Subtitle Generation Rules

**Input:** All audio files with speech marks

**Process:**
1. Iterate through scenes in order
2. For each scene, parse speech marks JSON
3. Generate subtitle entries with word-level timing
4. Output SRT format

**Subtitle Timing Rules:**
- Maximum line length: 42 characters
- Maximum 2 lines per subtitle
- Minimum duration: 1 second
- Maximum duration: 7 seconds
- Gap between subtitles: 100ms

**SRT Format:**
```
{subtitle_index}
{start_time} --> {end_time}
{subtitle_text}

{blank line}
```

Time format: `HH:MM:SS,mmm`

### 4.6 Video Rendering Rules

**FFmpeg Pipeline:**

1. **Per Scene Clip:**
   - Input: Image PNG, Audio MP3
   - Apply Ken Burns effect based on `camera_motion`
   - Duration matches audio length
   - Output: Individual scene clip

2. **Ken Burns Effects:**
   - `slow_zoom_in`: Scale from 1.0 to 1.1 over duration
   - `pan_left`: Pan from right to left
   - `pan_right`: Pan from left to right
   - `zoom_out`: Scale from 1.1 to 1.0

3. **Final Assembly:**
   - Concatenate scene clips in order
   - Overlay subtitles (SRT burn-in)
   - Add fade transitions between scenes
   - Export: H.264, 30 FPS, 2000k bitrate

**FFmpeg Command Example:**
```bash
# Individual scene with Ken Burns
ffmpeg -loop 1 -i scene.png -i scene.mp3 \
  -filter_complex "[0:v]scale=1080:1920,zoompan=z='zoom+0.001':d={duration}:s=1080x1920[v]" \
  -map "[v]" -map 1:a -t {duration} -c:v libx264 -preset fast \
  scene_clip.mp4

# Final assembly with subtitles
ffmpeg -f concat -i scenes.txt -vf subtitles=video.srt \
  -c:v libx264 -preset fast -b:v 2000k draft.mp4
```

### 4.7 Asset Completion Check

**Trigger:** After each `VOICE_GENERATION` or `IMAGE_GENERATION` job completes

**Logic:**
```python
scene_count = COUNT(scenes WHERE project_id = X)
audio_count = COUNT(audio_files WHERE project_id = X)
image_count = COUNT(assets WHERE project_id = X AND asset_type = 'IMAGE')

if audio_count >= scene_count AND image_count >= scene_count:
    project.status = 'ASSETS_READY'
    enqueue_job('SUBTITLE_GENERATION')
```

**Important:** This prevents race conditions when voice/image jobs complete out of order.

---

## 5. Frontend Specifications

### 5.1 Dashboard View

**Purpose:** List all projects and show their current status.

**Layout:**
- Table or card grid showing projects
- Columns: Title/Topic, Status Badge, Progress, Created Date, Actions
- Auto-refresh every 5 seconds when projects are processing
- Search/filter by status
- Sort by created date (newest first)

**Status Badge Colors:**
- **Blue**: Processing states (GENERATING, RENDERING)
- **Yellow**: Needs attention (IN_REVIEW)
- **Green**: Complete (APPROVED, RENDER_READY)
- **Red**: Error (FAILED, REJECTED)

**Actions:**
- Click row → Navigate to project detail
- "New Project" button → Create project form

### 5.2 Create Project Form

**Mode Toggle:** Radio buttons or tabs
- "Generate from Topic" (default)
- "Custom Script"

**Fields (Generate from Topic mode):**
- Topic: Text input, required
- Title: Text input, optional
- Target Duration: Dropdown (30s, 60s, 90s, 120s), default 60s
- Aspect Ratio: Dropdown (9:16, 16:9, 1:1), default 9:16
- Channel Style: Hidden or dropdown, default "dramatic_history_shorts"

**Fields (Custom Script mode):**
- Title: Text input, recommended
- Script: Textarea, required, min 50 characters
- Target Duration: Same as above
- Aspect Ratio: Same as above

**Submit:**
- Button: "Create Project"
- Loading state while API request pending
- On success: Redirect to project detail page
- On error: Show error message inline

### 5.3 Project Detail View

**Header:**
- Title (or topic if no title)
- Status badge
- Created date
- Current step indicator

**Progress Timeline:**
- Visual pipeline showing completed/current/upcoming steps
- Steps: Script → Scenes → Assets → Subtitles → Render → Review
- Checkmarks for completed, spinner for current, gray for pending

**Tabs:**
1. **Overview**
   - Project metadata
   - Processing log (completed steps)
   
2. **Script**
   - Show full script
   - For AI-generated: Show sections (hook, setup, etc.)
   - For custom: Show as single block

3. **Scenes**
   - Accordion or card list
   - Each scene shows:
     - Scene number and duration
     - Narration text
     - Mood and keywords
     - Visual prompt (collapsible)
     - Preview thumbnails (if available)

4. **Preview** (only if status >= RENDER_READY)
   - HTML5 video player
   - Download button
   - Approve/Reject buttons (if status == IN_REVIEW)

**Actions (conditional on status):**
- `IN_REVIEW`: Approve, Reject, Download
- `FAILED`: Retry
- `RENDER_READY` or `APPROVED`: Download

**Approve/Reject Modal:**
- Notes textarea (optional)
- Acted By field (email or name)
- Submit button

### 5.4 Video Player Requirements

**HTML5 Video Player:**
```html
<video controls width="360" height="640">
  <source src="/projects/{id}/download" type="video/mp4">
</video>
```

**Download Function:**
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

### 5.5 Polling Logic

**Auto-refresh for processing projects:**
```javascript
useEffect(() => {
  const interval = setInterval(() => {
    if (isProcessing(project.status)) {
      refetchProject();
    }
  }, 5000); // Every 5 seconds
  
  return () => clearInterval(interval);
}, [project.status]);

function isProcessing(status) {
  return [
    'CREATED',
    'SCRIPT_GENERATING',
    'SCENES_GENERATING',
    'VOICE_GENERATING',
    'ASSETS_GENERATING',
    'SUBTITLES_GENERATING',
    'RENDERING'
  ].includes(status);
}
```

### 5.6 Error Handling

**Toast Notifications:**
- Success: "Project created successfully"
- Error: "Failed to create project: {error message}"
- Info: "Project approved"

**Inline Errors:**
- Form validation errors below fields
- API errors in red alert box

**Project Failed State:**
- Show error message from `project.error_message`
- Retry button to call `/projects/{id}/retry`
- Error details expandable section

---

## 6. Job Processing Specifications

### 6.1 Job Lifecycle

```
QUEUED → RUNNING → SUCCEEDED
                 ↓
                FAILED → RETRYING (attempt < max_attempts)
                      ↓
                     FAILED (attempt >= max_attempts, permanent)
```

### 6.2 Retry Logic

**Retry Policy:**
- Maximum attempts: 3 (configurable per job)
- Backoff: Exponential (1s, 2s, 4s)
- Retryable errors: Transient failures (network, provider rate limits)
- Non-retryable errors: Invalid input, authentication failure

**Permanent Failure:**
- When `attempt_count >= max_attempts`
- Callback: `workflow.MarkProjectFailed(project_id, error_message)`
- Project status set to `FAILED`
- Error message stored in `projects.error_message`

### 6.3 Handler Implementation Pattern

```go
func HandleJobType(ctx context.Context, job *models.Job) error {
    // 1. Parse payload
    payload, err := jobs.ParsePayload[JobTypePayload](job)
    if err != nil {
        return err // Non-retryable
    }
    
    // 2. Execute business logic
    result, err := service.DoWork(ctx, payload)
    if err != nil {
        return err // Retryable
    }
    
    // 3. Call workflow callback
    return workflow.OnJobCompleted(ctx, payload.ProjectID, result)
}
```

### 6.4 Queue Backend Comparison

| Feature | Memory Queue | Redis Queue |
|---------|--------------|-------------|
| Persistence | DB only | DB + Redis |
| Multi-worker | No | Yes |
| Scalability | Single process | Horizontal |
| Dev/Test | ✅ Recommended | Overkill |
| Production | ❌ Not recommended | ✅ Recommended |

---

## 7. Provider Specifications

### 7.1 Voice Provider Interface

```go
type Provider interface {
    Generate(ctx context.Context, input GenerateInput) (*GeneratedAudio, error)
}

type GenerateInput struct {
    Text     string
    VoiceID  string
    Engine   string // "standard" or "neural"
    TextType string // "text" or "ssml"
}

type GeneratedAudio struct {
    AudioData    []byte
    DurationMs   int
    SpeechMarks  []SpeechMark
}

type SpeechMark struct {
    Time  int    // Milliseconds
    Type  string // "word", "sentence", "viseme"
    Value string
}
```

### 7.2 Image Provider Interface

```go
type Provider interface {
    Generate(ctx context.Context, input GenerateInput) (*GeneratedImage, error)
}

type GenerateInput struct {
    Prompt         string
    NegativePrompt string
    AspectRatio    string
    StyleProfile   string
}

type GeneratedImage struct {
    ImageData []byte
    Format    string // "png", "jpg"
    Width     int
    Height    int
    Metadata  map[string]interface{}
}
```

### 7.3 LLM Provider Interface

```go
type ScriptGenerator interface {
    GenerateScript(ctx context.Context, input ScriptGenerateInput) (*Script, error)
}

type SceneGenerator interface {
    GenerateScenes(ctx context.Context, input SceneGenerateInput) ([]*Scene, error)
}
```

---

## 8. Performance Targets

| Metric | Mock Mode | Production (with AI) |
|--------|-----------|---------------------|
| Script generation | <1s | 10-30s |
| Scene generation | <1s | 10-30s |
| Voice per scene | <1s | 2-5s |
| Image per scene | <1s | 10-20s |
| Subtitle generation | <1s | <1s |
| Video render (8 scenes) | <5s | 15-30s |
| **Total pipeline** | **5-10s** | **2-5 minutes** |

### Scaling Recommendations
- Worker count: 3-5 workers for typical load
- Redis queue: Required for multi-worker setups
- Database: Connection pool size = worker_count * 2
- Storage: S3 recommended for production (faster than local filesystem)

---

## 9. Security Considerations

### 9.1 Authentication (Not Implemented)
- **Current:** No authentication
- **Production Requirement:** Add JWT or API key authentication middleware
- **Recommended:** OAuth2 with Google/GitHub for user login

### 9.2 Authorization (Not Implemented)
- **Current:** No project ownership checks
- **Production Requirement:** User can only access their own projects
- **Implementation:** Add `user_id` field to projects table

### 9.3 Input Validation
- **Implemented:** Basic validation on API requests
- **Required:** Sanitize user input for SQL injection (handled by pgx parameterized queries)
- **Required:** Validate file uploads (future feature)

### 9.4 API Rate Limiting (Not Implemented)
- **Production Requirement:** Rate limit project creation (e.g., 10 per hour per user)
- **Implementation:** Redis-based rate limiter

### 9.5 Secrets Management
- **Current:** Environment variables
- **Production:** Use AWS Secrets Manager or HashiCorp Vault
- **Never commit:** `.env` files with real credentials

---

## 10. Deployment Specifications

### 10.1 Docker Compose (Development/Demo)

**Services:**
- postgres: PostgreSQL 16
- redis: Redis 7
- api: API server (port 8080)
- worker: Job worker (3 instances)
- frontend: React app (port 3000)

**Volumes:**
- postgres_data: Database persistence
- redis_data: Queue persistence
- storage_data: Shared file storage

### 10.2 Production Deployment (AWS Example)

**Architecture:**
- ECS Fargate: API and Worker containers
- RDS PostgreSQL: Database
- ElastiCache Redis: Job queue
- S3: Media storage
- CloudFront: CDN for video delivery
- ALB: Load balancer for API
- ECR: Container registry

**Environment Variables:**
```bash
DATABASE_URL=postgres://...rds.amazonaws.com/youtube_automation
REDIS_URL=redis://...cache.amazonaws.com:6379
STORAGE_BACKEND=s3
STORAGE_BUCKET=prod-youtube-automation
VOICE_PROVIDER=polly
IMAGE_PROVIDER=openai
LLM_PROVIDER=openai
QUEUE_BACKEND=redis
WORKER_COUNT=5
USE_MOCK_RENDER=false
```

### 10.3 Monitoring Requirements

**Metrics:**
- Job queue depth
- Job processing time per type
- Failed job rate
- Project completion rate
- Storage usage

**Logging:**
- Structured JSON logs
- Log aggregation (CloudWatch, DataDog, etc.)
- Include `project_id` in all logs

**Alerting:**
- Job failure rate > 10%
- Queue depth > 100
- Database connection errors
- Storage quota exceeded

---

This specification document provides complete details for understanding and implementing the YouTube Video Automation System, covering business logic, data models, APIs, processing rules, and deployment considerations.
