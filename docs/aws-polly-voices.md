# AWS Polly Voice Options

When creating a project, you can specify which AWS Polly voice to use via the `voice_id` and `voice_engine` fields.

## Voice Engines

- **standard**: Traditional TTS, good quality, lower cost
- **neural**: AI-enhanced, more natural sounding, higher cost (recommended for most use cases)
- **long-form**: Optimized for longer content (3+ minutes), consistent quality throughout, best for extended narration
- **generative**: Latest AI model, most natural and expressive, highest cost (recommended for premium content)

**Note**: Not all voices support all engines. Generative and long-form engines are available for select voices only.

## Available Voices

### English (US)

#### Long-form Voices (Best for Extended Narration)
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Matthew` | Male | Clear, authoritative - excellent for documentaries |
| `Joanna` | Female | Warm, professional - excellent for storytelling |
| `Ruth` | Female | News anchor style |
| `Stephen` | Male | Conversational, professional |

**Best for**: Videos longer than 3 minutes, audiobooks, documentaries, extended educational content

#### Generative Voices (Latest - Most Natural)
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Ruth` | Female | Conversational, news-style, highly expressive |
| `Stephen` | Male | Conversational, news-style, highly expressive |
| `Matthew` | Male | Clear, authoritative (also supports neural) |
| `Joanna` | Female | Warm, professional (also supports neural) |

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Joanna` | Female | Warm, professional, news anchor style |
| `Matthew` | Male | Clear, authoritative, news anchor style |
| `Ivy` | Female (child) | Young, friendly |
| `Kendra` | Female | Neutral, professional |
| `Kimberly` | Female | Warm, conversational |
| `Salli` | Female | Clear, friendly |
| `Joey` | Male | Casual, friendly |
| `Justin` | Male (child) | Young, energetic |
| `Kevin` | Male (child) | Young, friendly |
| `Ruth` | Female | Professional, clear |
| `Stephen` | Male | Professional, authoritative |

#### Standard Voices
All neural voices above also support standard engine, plus:
- `Danielle` (Female)
- `Gregory` (Male)

### English (UK)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Amy` | Female | Clear, British accent |
| `Emma` | Female | Warm, British accent |
| `Brian` | Male | Professional, British accent |
| `Arthur` | Male | Professional, British accent |

#### Standard Voices
- `Amy`, `Emma`, `Brian` (also support neural)
- `Geraint` (Male, Welsh)

### English (Australian)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Nicole` | Female | Clear, Australian accent |
| `Olivia` | Female | Warm, Australian accent |
| `Russell` | Male | Professional, Australian accent |

### English (Indian)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Kajal` | Female | Clear, Indian accent |
| `Raveena` | Female | Warm, Indian accent (standard only) |

### English (South African)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Ayanda` | Female | Clear, South African accent |

### English (New Zealand)

#### Standard Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Aria` | Female | New Zealand accent |

### Spanish (US)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Lupe` | Female | Clear, Latin American Spanish |
| `Pedro` | Male | Professional, Latin American Spanish |

### Spanish (European)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Lucia` | Female | Clear, Castilian Spanish |
| `Sergio` | Male | Professional, Castilian Spanish |

### Spanish (Mexican)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Mia` | Female | Clear, Mexican Spanish |
| `Andres` | Male | Professional, Mexican Spanish |

### French

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Lea` | Female | Clear, French accent |
| `Remi` | Male | Professional, French accent |

### French (Canadian)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Gabrielle` | Female | Clear, Quebec French |
| `Liam` | Male | Professional, Quebec French |

### German

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Vicki` | Female | Clear, German accent |
| `Daniel` | Male | Professional, German accent |
| `Hannah` | Female | Warm, Austrian German |

### Italian

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Bianca` | Female | Clear, Italian accent |
| `Adriano` | Male | Professional, Italian accent |

### Portuguese (Brazilian)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Camila` | Female | Clear, Brazilian accent |
| `Vitoria` | Female | Warm, Brazilian accent |
| `Thiago` | Male | Professional, Brazilian accent |

### Portuguese (European)

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Ines` | Female | Clear, European Portuguese |

### Japanese

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Takumi` | Male | Professional, Japanese |
| `Kazuha` | Female | Clear, Japanese |
| `Tomoko` | Female | Warm, Japanese (standard only) |

### Korean

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Seoyeon` | Female | Clear, Korean |

### Mandarin Chinese

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Zhiyu` | Female | Clear, Mandarin |

### Arabic

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Hala` | Female | Clear, Gulf Arabic |
| `Zayd` | Male | Professional, Gulf Arabic |

### Hindi

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Aditi` | Female | Clear, Hindi |

### Turkish

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Filiz` | Female | Clear, Turkish |

### Polish

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Ola` | Female | Clear, Polish |

### Dutch

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Laura` | Female | Clear, Dutch |

### Norwegian

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Ida` | Female | Clear, Norwegian |

### Swedish

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Elin` | Female | Clear, Swedish |

### Danish

#### Neural Voices
| Voice ID | Gender | Description |
|----------|--------|-------------|
| `Sofie` | Female | Clear, Danish |

## Usage in API

### Create Project with Custom Voice

```json
{
  "topic": "The Rise and Fall of Mansa Musa",
  "title": "Mansa Musa: The Richest Man in History",
  "channel_style": "dramatic_history_shorts",
  "target_duration_sec": 120,
  "aspect_ratio": "9:16",
  "voice_id": "Matthew",
  "voice_engine": "generative"
}
```

**Valid engine values**: `"standard"`, `"neural"`, or `"generative"`

### Default Behavior

If `voice_id` and `voice_engine` are not specified:
- **voice_id**: `Ayanda` (South African English, female)
- **voice_engine**: `standard`

These defaults can be changed via environment variables:
- `POLLY_DEFAULT_VOICE`
- `POLLY_ENGINE`

## Recommendations for History Content

### Long Documentary (5+ minutes) - Long-form Engine
- **Male voices**: `Matthew` (long-form) - Best for extended history narratives
- **Female voices**: `Joanna` (long-form), `Ruth` (long-form)
- **Why**: Maintains consistent quality and tone throughout longer content

### Premium Short-form (1-3 minutes) - Generative Engine
- **Male voices**: `Stephen` (generative), `Matthew` (generative)
- **Female voices**: `Ruth` (generative), `Joanna` (generative)
- **Why**: Most natural and engaging for shorter viral content

### Standard Production (Neural Engine)
- **Male voices**: `Matthew` (neural), `Brian` (neural, UK), `Russell` (neural, Australian)
- **Female voices**: `Joanna` (neural), `Amy` (neural, UK), `Ayanda` (neural, South African)

### Documentary Style
- **Male voices**: `Stephen` (neural), `Arthur` (neural, UK)
- **Female voices**: `Ruth` (neural), `Kendra` (neural)

### Educational/Friendly
- **Male voices**: `Joey` (neural), `Justin` (neural, young)
- **Female voices**: `Salli` (neural), `Kimberly` (neural)

## Cost Considerations

**Pricing by engine** (per 1 million characters):
- **Generative**: $30 (most natural and expressive)
- **Long-form**: $100 (optimized for consistency in longer content)
- **Neural**: $16 (great quality, best value)
- **Standard**: $4 (basic quality)

For a typical 120-second video (~300 words = ~1,800 characters):
- **Generative**: ~$0.054 per video
- **Long-form**: ~$0.180 per video
- **Neural**: ~$0.029 per video
- **Standard**: ~$0.007 per video

**Important**: Long-form pricing is significantly higher but provides better consistency for content over 3 minutes.

**Cost vs Quality Trade-off**:
- Use **long-form** for videos over 3 minutes where consistency matters (documentaries, audiobooks)
- Use **generative** for premium short-form content (under 3 minutes)
- Use **neural** for most production content (best balance of cost/quality)
- Use **standard** for testing, drafts, or high-volume low-budget content

## Testing Voices

To test different voices, create projects with different `voice_id` values and compare the output quality.

## Notes

1. **Long-form engine is designed for 3+ minute content** - provides better consistency for extended narration
2. **Generative engine is newest and most natural** for short-form - best for flagship content under 3 minutes
3. **Neural engine is recommended** for most production videos - great balance of cost and quality
4. **Not all voices support all engines**: 
   - Long-form: Matthew, Joanna, Ruth, Stephen (US English)
   - Generative: Ruth, Stephen, Matthew, Joanna (US English)
   - Neural: Most modern voices support it
   - Standard: All voices support it
4. **Voice consistency**: Use the same voice AND engine across all videos in a channel for brand consistency
5. **Accent matching**: Choose voices that match your target audience's region
6. **Character limits**: AWS Polly has a limit of 3,000 characters per request (the system automatically handles this)
7. **Testing**: Try generative engine on one video to compare quality before committing to higher costs

## Additional Resources

- [AWS Polly Voice List](https://docs.aws.amazon.com/polly/latest/dg/voicelist.html)
- [AWS Polly Pricing](https://aws.amazon.com/polly/pricing/)
- [SSML Support](https://docs.aws.amazon.com/polly/latest/dg/supportedtags.html) (for advanced control)
