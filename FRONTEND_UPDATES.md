# Frontend Updates for Voice Selection

## Summary

The frontend has been updated to support the new voice selection features added to the backend.

## Changes Made

### File: `frontend/src/components/CreateProjectForm.tsx`

#### 1. New State Variables

Added two new state variables for voice settings:

```typescript
const [voiceId, setVoiceId] = useState('Matthew');
const [voiceEngine, setVoiceEngine] = useState('neural');
```

**Defaults:**
- Voice: `Matthew` (US English, male, authoritative)
- Engine: `neural` (recommended quality/cost balance)

#### 2. Updated API Payload

Both topic and script modes now include voice settings in the API request:

```typescript
{
  topic: topic.trim(),
  title: title.trim(),
  channel_style: 'dramatic_history_shorts',
  target_duration_sec: duration,
  aspect_ratio: aspectRatio,
  voice_id: voiceId,           // NEW
  voice_engine: voiceEngine,    // NEW
}
```

#### 3. New Voice Settings Section

Added a complete voice settings section with:

**Voice Selection Dropdown:**
- Organized by language/region groups
- 25+ popular voices included
- Grouped into:
  - US English (Male): Matthew, Joey, Justin, Stephen
  - US English (Female): Joanna, Kendra, Kimberly, Salli, Ruth, Ivy
  - UK English: Brian, Amy, Emma, Arthur
  - Australian English: Russell, Nicole, Olivia
  - Other English: Ayanda (South African), Kajal (Indian)

**Engine/Quality Dropdown:**
- `generative` - Best Quality (Premium)
- `neural` - Great Quality (Recommended) ⭐
- `standard` - Basic Quality

**Dynamic Hints:**
- Shows context-aware hints based on selected engine
- Explains quality/cost trade-offs
- Warns about generative engine voice limitations

#### 4. Updated Aspect Ratio Options

Added 1:1 square format option:
- 9:16 Portrait (Shorts)
- 16:9 Landscape
- 1:1 Square (NEW)

## UI Preview

### Voice Settings Section

```
┌─────────────────────────────────────────────────┐
│ Voice Settings                                  │
├─────────────────────────────────────────────────┤
│ Voice                    Quality                │
│ ┌──────────────────┐    ┌──────────────────┐   │
│ │ Matthew          │    │ Neural (Great)   │   │
│ │   Authoritative  │    │   Quality        │   │
│ └──────────────────┘    └──────────────────┘   │
│                                                 │
│ ⭐ Recommended - Natural sounding with great   │
│    value.                                       │
└─────────────────────────────────────────────────┘
```

## Testing the Frontend

### 1. Start the Frontend

```bash
cd frontend
npm install  # If dependencies not installed
npm run dev
```

The app will be available at `http://localhost:5173`

### 2. Test Voice Selection

1. Navigate to "New Project" page
2. Fill in topic (e.g., "The Fall of Rome")
3. Select a voice from dropdown (e.g., "Ruth - News Anchor")
4. Select engine quality (e.g., "Generative")
5. Click "Create Project"

### 3. Verify API Call

Open browser DevTools > Network tab and verify the POST request includes:

```json
{
  "topic": "The Fall of Rome",
  "voice_id": "Ruth",
  "voice_engine": "generative",
  "aspect_ratio": "16:9",
  "target_duration_sec": 60
}
```

## Voice Recommendations in UI

The form provides contextual hints:

### When Generative Selected
> 🎯 Premium quality - Most natural. Only available for Ruth, Stephen, Matthew, Joanna.

**Warning**: If user selects a voice that doesn't support generative (e.g., "Brian"), the backend will fail. Consider adding validation.

### When Neural Selected
> ⭐ Recommended - Natural sounding with great value.

### When Standard Selected
> 💰 Basic quality - Good for testing.

## Future Enhancements

### 1. Voice Preview
Add audio samples for each voice:

```typescript
<button onClick={() => playVoicePreview(voiceId)}>
  🔊 Preview
</button>
```

### 2. Engine Compatibility Validation

Warn users when selecting incompatible voice/engine combinations:

```typescript
const generativeVoices = ['Ruth', 'Stephen', 'Matthew', 'Joanna'];

if (voiceEngine === 'generative' && !generativeVoices.includes(voiceId)) {
  // Show warning or auto-change to neural
}
```

### 3. Cost Estimator

Show estimated voice generation cost:

```typescript
const estimateCost = (duration: number, engine: string) => {
  const chars = duration * 10; // ~10 chars per second
  const rates = { standard: 0.000004, neural: 0.000016, generative: 0.00003 };
  return (chars * rates[engine]).toFixed(4);
};

<p className="form-hint">
  Estimated voice cost: ${estimateCost(duration, voiceEngine)}
</p>
```

### 4. Filter Voices by Language

Add language filter dropdown:

```typescript
const [language, setLanguage] = useState('en-US');

<select value={language} onChange={e => setLanguage(e.target.value)}>
  <option value="en-US">US English</option>
  <option value="en-GB">UK English</option>
  <option value="en-AU">Australian English</option>
</select>
```

### 5. Save Voice Preferences

Remember user's last selected voice:

```typescript
// On submit
localStorage.setItem('preferredVoice', voiceId);
localStorage.setItem('preferredEngine', voiceEngine);

// On component mount
useEffect(() => {
  const savedVoice = localStorage.getItem('preferredVoice');
  if (savedVoice) setVoiceId(savedVoice);
  
  const savedEngine = localStorage.getItem('preferredEngine');
  if (savedEngine) setVoiceEngine(savedEngine);
}, []);
```

## CSS Considerations

The form uses existing CSS classes. If you want custom styling for voice section:

```css
.voice-settings {
  background: var(--bg-secondary);
  padding: 1rem;
  border-radius: 8px;
  margin-top: 1rem;
}

.voice-hint-premium {
  color: var(--accent);
  font-weight: 500;
}

.voice-hint-recommended {
  color: var(--success);
}

.voice-hint-basic {
  color: var(--text-secondary);
}
```

## Accessibility

Ensure proper labels and ARIA attributes:

```typescript
<label className="form-label" htmlFor="field-voice-id">
  Voice
  <span className="sr-only">Choose text-to-speech voice</span>
</label>

<select
  id="field-voice-id"
  aria-describedby="voice-hint"
  ...
>
```

## API Integration

The form uses the existing `createProject` function from `api.ts`. No changes needed to the API client - it automatically includes the new fields.

## Backward Compatibility

**Existing projects**: The backend will use defaults if `voice_id` and `voice_engine` are not provided, so existing frontend code will continue to work.

**New frontend with old backend**: If the backend hasn't been updated, the fields will be ignored (no errors).

## Migration Checklist

- [x] Update `CreateProjectForm.tsx` with voice fields
- [x] Add voice selection dropdown (25+ voices)
- [x] Add engine quality dropdown (standard/neural/generative)
- [x] Add contextual hints for each engine
- [x] Update aspect ratio to include 1:1 square
- [x] Ensure API payload includes new fields
- [ ] Test with backend API
- [ ] Add voice compatibility validation (future)
- [ ] Add voice preview feature (future)
- [ ] Add cost estimator (future)

## Testing Checklist

### Functionality
- [ ] Form submits with voice settings
- [ ] All voice options are selectable
- [ ] All engine options are selectable
- [ ] Hints update when engine changes
- [ ] Project creates successfully with selected voice
- [ ] Generated audio uses selected voice

### UI/UX
- [ ] Voice dropdown is readable
- [ ] Grouped voices are clearly organized
- [ ] Hints are visible and helpful
- [ ] Form layout looks good on mobile
- [ ] Form layout looks good on desktop

### Edge Cases
- [ ] What happens if voice doesn't support generative?
- [ ] Does form handle very long voice names?
- [ ] Does form validate voice/engine combination?

## Known Limitations

1. **No Voice Preview**: Users can't hear voice samples before selection
2. **No Validation**: Form doesn't prevent selecting incompatible voice/engine combos
3. **No Cost Display**: Users don't see estimated costs upfront
4. **Limited Voices**: Only ~25 voices shown (AWS has 60+)
5. **No Language Filter**: All voices shown together (could be overwhelming)

## Summary

✅ Frontend now supports:
- Voice selection (25+ voices)
- Engine quality selection (standard/neural/generative)
- Aspect ratio updates (added 1:1 square)
- Contextual hints and guidance

Users can now fully customize voice settings when creating projects!
