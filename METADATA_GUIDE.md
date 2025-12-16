# Content Metadata & Multi-Platform Distribution Guide

## Overview

This system generates platform-optimized metadata for all your content in a **single LLM call**, maximizing token efficiency while ensuring each platform gets perfectly tailored metadata for maximum organic reach.

## What Gets Generated

For each news article, the system creates:

1. **News Summary** (150-200 words, broadcast-quality)
2. **SEO Metadata** (keywords, topics, sentiment, target audience)
3. **Platform-Specific Metadata** for 6 platforms:
   - YouTube (title, description, tags, timestamps)
   - TikTok (caption, hashtags)
   - Instagram (caption, hashtags)
   - Twitter (tweet, hashtags)
   - Facebook (post text, link description)
   - LinkedIn (professional post, hashtags)

All of this is generated in **one LLM API call** for maximum efficiency.

## Data Storage

### Content Manifest File

All content and metadata is stored in `content_manifest.json`:

```json
{
  "version": "1.0",
  "generated_at": "2025-12-13T10:00:00Z",
  "items": [
    {
      "id": "news_20251213_100000",
      "created_at": "2025-12-13T10:00:00Z",
      "source": {
        "url": "https://...",
        "title": "Original Article Title",
        "source_name": "cnn.com"
      },
      "content": {
        "summary": "150-200 word broadcast summary...",
        "word_count": 175,
        "estimated_duration_seconds": 60
      },
      "media": {
        "video_path": "news_20251213_100000_final.mp4",
        "duration_seconds": 62.5,
        "resolution": "1280x720",
        "avatar_id": "Kristin_public_3_20240108"
      },
      "seo": {
        "primary_keywords": ["AI", "technology", "innovation"],
        "topics": ["Technology", "Business"],
        "sentiment": "positive"
      },
      "platforms": {
        "youtube": {
          "title": "Breaking: Major AI Breakthrough",
          "description": "Detailed YouTube description...",
          "tags": ["AI", "tech", "news"],
          "timestamps": [...]
        },
        "tiktok": {
          "caption": "🚨 Breaking AI news!",
          "hashtags": ["AI", "TechNews", "FYP"]
        },
        // ... other platforms
      },
      "posting_status": {
        "youtube": {"posted": false},
        "tiktok": {"posted": false}
        // ... other platforms
      }
    }
  ]
}
```

## Running the Pipeline

### Basic Usage

```bash
# Run the complete workflow
go run main.go

# What happens:
# 1. Fetches news article from NewsAPI
# 2. Generates summary + metadata (single LLM call)
# 3. Creates audio narration (Google TTS)
# 4. Generates AI avatar video (HeyGen)
# 5. Saves everything to content_manifest.json
```

### Output Files

After running, you'll have:

```
news_20251213_100000_final.mp4   ← Video ready to post
content_manifest.json              ← All metadata for posting
```

## Inspecting Metadata

Use the manifest inspection tool:

```bash
# List all content items
go run tools/inspect_manifest.go list

# Show full details for an item
go run tools/inspect_manifest.go show news_20251213_100000

# View platform-specific metadata
go run tools/inspect_manifest.go youtube news_20251213_100000
go run tools/inspect_manifest.go tiktok news_20251213_100000
go run tools/inspect_manifest.go instagram news_20251213_100000

# Show statistics
go run tools/inspect_manifest.go stats

# List unposted content
go run tools/inspect_manifest.go unposted
```

## Platform Optimization Details

### YouTube Optimization

**Goal:** Search visibility, watch time, recommendations

**Generated Metadata:**
- **Title** (100 chars max): Front-loads keywords, creates curiosity
- **Description** (5000 chars max): First 150 chars critical, includes keywords naturally
- **Tags** (3-5 optimal): Highly relevant, not generic
- **Timestamps**: Improves watch time and user experience
- **Category**: Auto-set to "News & Politics" (25)

**Example:**
```
Title: "Breaking: Major AI Breakthrough Could Change Everything"
Description: "A groundbreaking AI development was announced today...
0:00 Introduction
0:15 Key Development
0:30 Impact Analysis
0:45 What This Means For You"
Tags: ["AI breakthrough", "technology news", "artificial intelligence"]
```

### TikTok Optimization

**Goal:** Virality, trending, FYP (For You Page)

**Generated Metadata:**
- **Caption** (150 chars recommended): Hook + value + CTA
- **Hashtags** (3-5 optimal): Mix of trending + niche
- **Privacy**: Public by default
- **Duet/Stitch**: Enabled for engagement

**Example:**
```
Caption: "🚨 This AI news will blow your mind! Here's what you need to know 👇 #AI #TechNews #Breaking"
Hashtags: ["AI", "TechNews", "Breaking", "Innovation", "FYP"]
```

### Instagram Reels Optimization

**Goal:** Community engagement, hashtag discovery

**Generated Metadata:**
- **Caption**: Emoji-rich, paragraph breaks
- **Hashtags** (5-10 optimal): Mix of popular + niche
- **Location**: Optional (added if relevant to story)

**Example:**
```
Caption: "🔥 BREAKING AI NEWS 🔥

This just dropped and it's huge! Here's what everyone's talking about...

📌 Save this for later
👉 Follow for more tech updates

#AI #Technology #Innovation #TechNews #Reels"
```

### Twitter/X Optimization

**Goal:** Viral potential, conversation starter

**Generated Metadata:**
- **Tweet** (280 chars): Punchy, question or hot take
- **Hashtags** (1-2 max): More hurt engagement
- **Format**: Often thread starter format

**Example:**
```
Tweet: "🚨 BREAKING: Major AI breakthrough just announced

This could change everything we know about technology.

Here's what you need to know 🧵👇"
Hashtags: ["AI", "TechNews"]
```

### Facebook Optimization

**Goal:** Emotional engagement, shares, comments

**Generated Metadata:**
- **Post Text**: Longer-form (300-500 words performs well)
- **Emotional Hook**: Surprise, curiosity, concern
- **Questions**: Drive comments

**Example:**
```
Message: "This AI announcement has everyone talking, and for good reason...

[Detailed explanation with emotional hooks and questions]

What do you think about this development? Let me know in the comments!"
```

### LinkedIn Optimization

**Goal:** Thought leadership, professional credibility

**Generated Metadata:**
- **Post Text**: Professional tone, industry insights
- **Hot Takes**: Industry trend observations
- **Hashtags**: 3-5 professional hashtags

**Example:**
```
Post: "Industry Alert: Major AI Development

As business leaders, this announcement has significant implications for how we approach technology strategy in 2025.

Here's my analysis of what this means for your organization...

#ArtificialIntelligence #TechnologyLeadership #BusinessInnovation"
```

## Using Metadata for Posting

### Manual Posting

1. Generate content: `go run main.go`
2. Inspect metadata: `go run tools/inspect_manifest.go youtube <id>`
3. Copy metadata to platform
4. Upload video manually
5. Update posting status (future feature)

### Programmatic Posting (Future)

```go
// Example: Post to YouTube (future implementation)
import "content-generation-automation/platforms/youtube"

item, _ := manifestManager.GetItem("news_20251213_100000")
url, err := youtube.PostVideo(
    item.Media.VideoPath,
    item.Platforms.YouTube,
)

// Update posting status
item.Status.YouTube.Posted = true
item.Status.YouTube.URL = &url
manifestManager.UpdateItem(item)
```

## Token Efficiency Strategy

### Why Single LLM Call?

**Old Approach (inefficient):**
```
Call 1: Generate summary (2000 tokens)
Call 2: Generate YouTube metadata (500 tokens)
Call 3: Generate TikTok metadata (500 tokens)
Call 4: Generate Instagram metadata (500 tokens)
... (6 separate calls)
Total: ~4000 tokens + 6× API overhead
```

**New Approach (optimized):**
```
Call 1: Generate summary + all platform metadata (3000 tokens)
Total: ~3000 tokens + 1× API overhead
Savings: 25% fewer tokens, 6× fewer API calls
```

### Cost Comparison

Per article with Grok-2:
- **Old method**: 6 API calls × $0.00003 = $0.00018
- **New method**: 1 API call × $0.00003 = $0.00003
- **Savings**: 83% reduction in API overhead

At scale (1000 articles/month):
- **Old method**: $180/month
- **New method**: $30/month
- **Annual savings**: $1,800

## Customizing Metadata Generation

### Modify the Prompt

Edit `metadata/prompts.go` to customize:

```go
func MetadataGenerationPrompt() string {
    return `Your custom prompt here...

    Customize:
    - Tone and voice
    - Hashtag strategy
    - Keyword focus
    - Platform priorities
    - Brand guidelines
    `
}
```

### Adjust Platform Settings

Edit the conversion function in `news/metadata_generation.go`:

```go
YouTube: metadata.YouTubeMetadata{
    CategoryID: "28",  // Change category
    PrivacyStatus: "unlisted",  // Change privacy
    // ... other settings
},
```

## Best Practices

### Content Strategy

1. **Generate in batches** - Process multiple articles at once
2. **Review metadata** - Always inspect before posting
3. **Track performance** - Monitor which metadata performs best
4. **A/B test** - Try different title/caption strategies
5. **Update templates** - Refine prompt based on what works

### Posting Strategy

1. **Platform priority**: YouTube → TikTok → Instagram → Twitter → LinkedIn → Facebook
2. **Timing**: Post during peak hours for each platform
3. **Engagement**: Respond to comments quickly
4. **Cross-promote**: Mention other platforms in posts
5. **Analytics**: Track which platforms drive most revenue

### Monetization Strategy

- **YouTube**: Ad revenue + memberships
- **TikTok**: Creator fund + brand deals
- **Instagram**: Sponsored posts + affiliate links
- **Twitter**: Super Follows + tips
- **Facebook**: Ad revenue
- **LinkedIn**: Thought leadership → consulting clients

## Troubleshooting

### "Failed to parse LLM JSON response"

**Cause:** LLM returned invalid JSON or wrapped in markdown

**Solution:** The system automatically strips markdown code blocks, but if errors persist:
1. Check `metadata/prompts.go` emphasizes JSON output
2. Review raw LLM response in logs
3. Try different model (grok-2-1212 works well)

### Empty or Invalid Metadata

**Cause:** LLM couldn't access article or article was invalid

**Solution:**
1. Verify article URL is accessible
2. Check NewsAPI returned valid data
3. Review verification requirements in prompt

### Manifest File Corrupted

**Solution:**
```bash
# Backup current manifest
cp content_manifest.json content_manifest.backup.json

# Start fresh
rm content_manifest.json
go run main.go
```

## Future Enhancements

See `.cursor/scratchpad.md` for roadmap, including:

- [ ] Automated posting to all platforms
- [ ] A/B testing for metadata variations
- [ ] Performance analytics tracking
- [ ] SQLite database migration (for scale)
- [ ] Multi-language support
- [ ] Custom brand voice training
- [ ] Automated thumbnail generation
- [ ] Social media scheduling

## Support

For issues or questions:
1. Check this guide
2. Review `.cursor/scratchpad.md` for implementation details
3. Inspect manifest with CLI tools
4. Check logs for detailed error messages

---

**Built for maximum organic reach across all platforms! 🚀**

