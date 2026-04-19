# AWS Polly Voice Engine Comparison

## Quick Comparison

| Feature | Standard | Neural | Generative |
|---------|----------|--------|------------|
| **Quality** | Basic | Great | Exceptional |
| **Naturalness** | Robotic | Natural | Highly Natural |
| **Cost per 1M chars** | $4 | $16 | $30 |
| **Cost per 120s video** | $0.007 | $0.029 | $0.054 |
| **Availability** | All voices | Most voices | Limited voices |
| **Best For** | Testing, drafts | Production content | Premium content |
| **Processing Speed** | Fast | Medium | Medium |

## Detailed Comparison

### Standard Engine

**Pros:**
- ✅ Lowest cost ($4 per 1M characters)
- ✅ Fast processing
- ✅ Works with all voices
- ✅ Good for testing and prototyping

**Cons:**
- ❌ Robotic sound
- ❌ Less natural prosody
- ❌ Limited emotional range
- ❌ Noticeable pauses between words

**Use Cases:**
- Testing scripts and timing
- Draft videos for internal review
- High-volume content where cost is critical
- Non-customer-facing content

**Example Request:**
```json
{
  "topic": "The Fall of Rome",
  "voice_id": "Matthew",
  "voice_engine": "standard"
}
```

---

### Neural Engine

**Pros:**
- ✅ Natural-sounding speech
- ✅ Good prosody and intonation
- ✅ Works with most modern voices
- ✅ Best cost/quality balance
- ✅ Smooth transitions between words

**Cons:**
- ❌ 4x more expensive than standard
- ❌ Not all voices supported
- ❌ Slightly slower than standard

**Use Cases:**
- Production YouTube videos
- Regular content releases
- Most customer-facing content
- **Recommended for 90% of use cases**

**Example Request:**
```json
{
  "topic": "The Fall of Rome",
  "voice_id": "Matthew",
  "voice_engine": "neural"
}
```

---

### Generative Engine (Latest)

**Pros:**
- ✅ Most natural and expressive
- ✅ Advanced emotional range
- ✅ News anchor quality
- ✅ Conversational and engaging
- ✅ Best for premium content

**Cons:**
- ❌ Most expensive (7.5x standard, 1.9x neural)
- ❌ Only available for select voices:
  - **US English**: Ruth, Stephen, Matthew, Joanna
- ❌ Limited language support currently

**Use Cases:**
- Flagship content
- Premium channels
- High-value videos (e.g., sponsored content)
- When voice quality is critical to brand
- Documentary-style narration

**Example Request:**
```json
{
  "topic": "The Fall of Rome",
  "voice_id": "Ruth",
  "voice_engine": "generative"
}
```

---

## Voice Support by Engine

### Generative Engine Voices (US English Only)

| Voice ID | Gender | Style |
|----------|--------|-------|
| `Ruth` | Female | Conversational, news-style |
| `Stephen` | Male | Conversational, news-style |
| `Matthew` | Male | Clear, authoritative |
| `Joanna` | Female | Warm, professional |

**Note**: These voices also support neural and standard engines.

### Neural Engine Voices

Most modern voices support neural:
- **US English**: Joanna, Matthew, Ivy, Kendra, Kimberly, Salli, Joey, Justin, Kevin, Ruth, Stephen
- **UK English**: Amy, Emma, Brian, Arthur
- **Australian**: Nicole, Olivia, Russell
- **And many more** (see `aws-polly-voices.md` for complete list)

### Standard Engine Voices

**All AWS Polly voices** support the standard engine.

---

## Cost Analysis

### Monthly Costs (Example Channel)

**Assumptions:**
- 30 videos per month
- 120 seconds per video
- ~1,800 characters per video

**Total characters per month**: 54,000 (0.054 million)

| Engine | Cost per Video | Monthly Cost (30 videos) | Annual Cost |
|--------|---------------|-------------------------|-------------|
| Standard | $0.007 | $0.22 | $2.60 |
| Neural | $0.029 | $0.86 | $10.37 |
| Generative | $0.054 | $1.62 | $19.44 |

**Break-even Analysis:**
- Difference neural vs standard: ~$8/year for 30 videos/month
- Difference generative vs neural: ~$9/year for 30 videos/month

**Conclusion**: The cost difference is minimal for most channels. Quality should be the primary decision factor.

---

## Listening Test Recommendations

Before committing to an engine, create test videos:

```bash
# Test all three engines with the same content
curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "The Viking Age",
    "voice_id": "Matthew",
    "voice_engine": "standard",
    "target_duration_sec": 60
  }'

curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "The Viking Age",
    "voice_id": "Matthew",
    "voice_engine": "neural",
    "target_duration_sec": 60
  }'

curl -X POST http://localhost:8080/projects \
  -H "Content-Type: application/json" \
  -d '{
    "topic": "The Viking Age",
    "voice_id": "Matthew",
    "voice_engine": "generative",
    "target_duration_sec": 60
  }'
```

Listen to all three and decide which quality level meets your needs.

---

## Recommendations by Channel Type

### Budget Channel (High Volume, Cost-Sensitive)
- **Engine**: Standard
- **Voice**: Any
- **Cost**: Lowest
- **Trade-off**: Accept robotic quality for cost savings

### Professional Channel (Most Common)
- **Engine**: Neural ⭐ **RECOMMENDED**
- **Voice**: Matthew, Joanna, Brian, Amy, etc.
- **Cost**: Moderate
- **Trade-off**: Best balance of cost and quality

### Premium Channel (Brand-Critical Content)
- **Engine**: Generative
- **Voice**: Ruth, Stephen, Matthew, Joanna
- **Cost**: Highest
- **Trade-off**: Pay premium for exceptional quality

### Documentary/Educational
- **Engine**: Neural or Generative
- **Voice**: Stephen (generative), Matthew (neural), Brian (neural)
- **Cost**: Moderate to High
- **Trade-off**: Quality matters for educational credibility

### Entertainment/Viral Content
- **Engine**: Neural
- **Voice**: Joey, Salli, Ivy (more casual voices)
- **Cost**: Moderate
- **Trade-off**: Natural sound without premium cost

---

## Migration Strategy

If upgrading from one engine to another:

### From Standard → Neural
1. Create 1-2 test videos with neural
2. Compare quality improvement
3. Update default in `.env`: `POLLY_ENGINE=neural`
4. Ensure budget can handle 4x cost increase

### From Neural → Generative
1. Verify your voice supports generative (Ruth, Stephen, Matthew, Joanna only)
2. Create test video
3. If quality improvement justifies 1.9x cost, update projects individually
4. Consider using generative only for flagship content

### Mixed Strategy (Recommended)
- **Generative**: 10% of content (flagship videos, premieres)
- **Neural**: 85% of content (regular uploads)
- **Standard**: 5% of content (testing, drafts)

This gives best quality where it matters while optimizing costs.

---

## Technical Notes

### Fallback Behavior

If a voice doesn't support the requested engine:
- AWS Polly will return an error
- The system will fail the job
- Solution: Use `standard` engine (works with all voices)

### Engine-Specific Features

**Generative Engine**:
- Better at handling complex sentences
- More expressive with punctuation
- Better pause timing

**Neural Engine**:
- Good with SSML tags
- Consistent quality across languages

**Standard Engine**:
- Most reliable
- Predictable output
- Fastest processing

---

## Future Considerations

AWS Polly may expand generative engine to:
- More languages
- More voices
- Additional features (tone control, style transfer)

Monitor AWS Polly announcements for updates.

---

## Summary Decision Matrix

| If You Want... | Choose |
|----------------|--------|
| Lowest cost | Standard |
| Best value | Neural ⭐ |
| Best quality | Generative |
| Works with all voices | Standard |
| Natural conversation style | Neural or Generative |
| News anchor quality | Generative (Ruth/Stephen) |
| To test quickly | Standard |
| Production content | Neural |
| Premium/flagship | Generative |

**Most Common Choice**: Neural engine with Matthew or Joanna voice.
