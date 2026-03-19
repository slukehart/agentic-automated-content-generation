# Metadata Generation Prompt

## Overview

This is the system prompt used in `metadata/prompts.go` → `MetadataGenerationPrompt()`. It is passed to Grok-3 as a single API call that generates the news summary and all platform metadata simultaneously.

The user prompt is constructed by `UserPromptForArticle(articleTitle, articleURL)` and includes the article title and URL. Grok is expected to fetch and read the article itself.

---

## What the System Prompt Asks For

The prompt instructs Grok to act as a "professional multi-platform content strategist and news writer" and to:

1. Visit the article URL
2. Read the full article content
3. Generate a broadcast-quality news summary (150–200 words)
4. Create platform-specific metadata for six platforms: YouTube, TikTok, Instagram, Twitter/X, Facebook, and LinkedIn

All of this is returned in a single JSON response — no markdown, no code blocks, raw JSON only.

---

## Expected JSON Output Schema

```json
{
  "summary": "150-200 word broadcast news summary",
  "seo": {
    "primary_keywords": ["keyword1", "keyword2", "keyword3"],
    "secondary_keywords": ["keyword4", "keyword5"],
    "topics": ["Topic1", "Topic2"],
    "sentiment": "positive|negative|neutral",
    "target_audience": ["audience1", "audience2"]
  },
  "platforms": {
    "youtube": {
      "title": "Engaging title (max 100 chars)",
      "description": "Detailed description with keywords (max 1000 chars recommended)",
      "tags": ["tag1", "tag2", "tag3"],
      "timestamps": [
        {"time": "0:00", "label": "Introduction"},
        {"time": "0:15", "label": "Key Point 1"},
        {"time": "0:30", "label": "Key Point 2"},
        {"time": "0:45", "label": "Conclusion"}
      ]
    },
    "tiktok": {
      "caption": "Hook + value + CTA (max 150 chars for best performance)",
      "hashtags": ["hashtag1", "hashtag2", "hashtag3"]
    },
    "instagram": {
      "caption": "Caption with emojis and line breaks (5-10 hashtags at end)",
      "hashtags": ["hashtag1", "hashtag2", "hashtag3"]
    },
    "twitter": {
      "tweet": "Concise hook with question or CTA (max 280 chars)",
      "hashtags": ["hashtag1", "hashtag2"]
    },
    "facebook": {
      "message": "Longer-form post with emotional hook",
      "link_description": "Preview text for link sharing"
    },
    "linkedin": {
      "post_text": "Professional tone, industry insights, thought leadership angle",
      "hashtags": ["ProfessionalHashtag1", "ProfessionalHashtag2"]
    }
  }
}
```

This maps directly to `metadata.LLMMetadataResponse` in `metadata/types.go`. After parsing, `ConvertToContentItem()` in `news/metadata_generation.go` enriches the struct with platform defaults (privacy status, category IDs, etc.) before writing to the manifest.

---

## Platform-Specific Formatting Rules

**YouTube**
- Title: front-load keywords, 60–70 chars optimal (max 100)
- Description: first 150 chars appear before "Show more" — make them count
- Max 1000 chars recommended (prompt notes YouTube does not afford more)
- Tags: 3–5 highly relevant, not generic
- Timestamps: improve watch time, use 4 markers (intro, 2 key points, conclusion)

**TikTok**
- Caption: hook in first 3 seconds of text, create FOMO, max 150 chars for performance
- Hashtags: 3–5 trending + niche (prompt discourages 30+)
- CTA: "Follow for more" or engagement prompt

**Instagram Reels**
- Caption: start with emoji, short paragraphs, first line hooks scrollers
- Hashtags: 5–10 mix of popular + niche (not maxing out at 30)
- Location tag included if relevant to the story

**Twitter/X**
- Tweet: question, stat, or provocative framing, max 280 chars
- Hashtags: 1–2 max (more hurt engagement per prompt guidance)
- Thread starter format suggested: "🧵 Here's what you need to know"

**Facebook**
- Emotional hook: surprise, curiosity, concern
- Longer content performs well (300–500 words)
- Ask questions to drive comments
- Paragraph breaks for readability

**LinkedIn**
- Professional tone, industry insights, thought leadership
- "Hot take" or industry trend as opener
- Tag relevant companies/people in post_text
- 3–5 professional hashtags

---

## Summary Requirements (from prompt)

- 150–200 words
- Neutral, factual, authoritative tone (broadcast style)
- Engaging hook to open
- All key points, context, background, verified facts
- Sufficient detail for ~60 seconds of spoken delivery
- Cite the original source
- No speculation, opinions, or emotional language
- Meaningful conclusion conveying the article's main point

---

## Fallback / Verification Behavior

If the article URL is inaccessible, Grok is instructed to:
1. Find the story from a reputable outlet (AP, Reuters, BBC, Guardian, Al Jazeera)
2. If no reliable source confirms the story, return the full JSON structure with empty strings and empty arrays

This means callers should validate that `summary` is non-empty before proceeding — which `GenerateEnrichedNewsContent()` already does.

---

## What Works Well

- Single-call efficiency: one API call returns summary + all six platforms' metadata + SEO info. This is explicitly noted as the "most token-efficient approach" in the code comments.
- The fallback instruction handles paywalled or broken URLs gracefully without requiring error-path logic in the caller.
- JSON-only output instruction reduces post-processing friction, though the caller does strip markdown code fences defensively.
- Platform character limits and hashtag counts are encoded directly in the prompt, keeping output immediately usable without normalization.
- Timestamp scaffolding (4 fixed markers) works well for short-form news videos where the structure is predictable.

## What Could Be Improved

- **Temperature 0.7 may introduce variability in structured output.** For JSON generation, 0.3–0.5 tends to produce more consistent formatting. Occasional JSON parse failures in `metadata_generation.go` suggest this is a live issue.
- **The prompt asks Grok to fetch a URL**, which means output quality is dependent on Grok's web access. If Grok cannot retrieve the page and the fallback search also fails, the caller gets an empty-string result rather than a clear error.
- **Timestamp times are hardcoded placeholders** (0:00, 0:15, 0:30, 0:45) in the schema example. Grok tends to echo these rather than generate meaningful markers, since it doesn't know the actual video duration at prompt time.
- **The YouTube description cap** is documented as 1000 chars in the prompt, but the `YouTubeMetadata` type comment says "Max 5000 chars." These are inconsistent; the actual YouTube API allows up to 5000 characters.
- **No output length hint per-platform.** For Facebook, the prompt says "300–500 words perform well" but the JSON field has no enforced length signal. In practice Grok sometimes under-generates for Facebook.
- **SEO `sentiment` field** is a free-form string. Constraining it to an enum (positive/negative/neutral) in the prompt is good, but the type is `string` — invalid values won't be caught at parse time.

## Related

- [[metadata-generation]] — Architecture of the metadata generation system
- [[grok-xai]] — API integration details
- [[summary-tuning]] — Related tuning notes for summary quality
