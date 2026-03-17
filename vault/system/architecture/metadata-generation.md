---
type: architecture
component: metadata
status: current
created: 2026-03-17
updated: 2026-03-17
tags: [metadata, grok, llm]
---

## Overview

The metadata-generation component makes two X.AI API calls per pipeline run: one to Grok (text) to generate a broadcast-quality news summary plus all six platform metadata sets in a single response, and one to Grok Image to generate a photorealistic newsroom background JPEG. Both functions live in `news/metadata_generation.go`.

## How It Works

### Text Metadata: `GenerateEnrichedNewsContent`

1. Reads `X_AI_KEY` from env
2. Creates a `grok.Client` via `github.com/SimonMorphy/grok-go` with a 5-minute timeout
3. Assembles two prompts:
   - System: `metadata.MetadataGenerationPrompt()` — instructs Grok to act as a content strategist, visit the URL, and return a specific JSON structure
   - User: `metadata.UserPromptForArticle(title, url)` — provides the article title and URL
4. Calls `grok.CreateChatCompletion` with model `grok-3`, temperature 0.7, max 4000 tokens
5. Strips any markdown code fences from the raw response string
6. Unmarshals JSON into `metadata.LLMMetadataResponse`
7. Returns `EnrichedNewsContent{Summary, Metadata}`

### Image Generation: `GenerateNewsroomBackground`

1. Reads `X_AI_KEY` from env
2. POST to `https://api.x.ai/v1/images/generations` with:
   - `model: "grok-2-image"`
   - `n: 1`
   - `response_format: "b64_json"`
   - A default prompt for a dark-blue photorealistic newsroom if none is provided
3. Decodes the base64 image data
4. Saves as `backgrounds/newsroom_{YYYYMMDD_HHMMSS}.jpg`
5. Returns `BackgroundImageResult{ImagePath}`

Note: Grok returns JPEG regardless of format requested.

## LLM Response Schema

The single Grok text call returns this JSON structure, parsed into `LLMMetadataResponse`:

```go
type LLMMetadataResponse struct {
    Summary   string                      `json:"summary"`
    SEO       SEOInfo                     `json:"seo"`
    Platforms LLMPlatformMetadataResponse `json:"platforms"`
}
```

`Platforms` contains `LLMYouTubeMetadata`, `LLMTikTokMetadata`, `LLMInstagramMetadata`, `LLMTwitterMetadata`, `LLMFacebookMetadata`, `LLMLinkedInMetadata` — all defined in `metadata/types.go`.

## System Prompt Details (from `metadata/prompts.go`)

- Instructs Grok to visit the article URL and read it
- Requests a 150–200 word broadcast-quality news summary (neutral, factual, ~60 seconds spoken)
- Requests platform-specific optimization for each of 6 platforms
- Mandates raw JSON output only — no markdown wrappers
- If article is inaccessible, instructs Grok to find the story from AP, Reuters, BBC, or Guardian
- Never fabricate information

## Key Files

- `news/metadata_generation.go` — `GenerateEnrichedNewsContent`, `GenerateNewsroomBackground`, `ConvertToContentItem`
- `metadata/prompts.go` — `MetadataGenerationPrompt()`, `UserPromptForArticle()`
- `metadata/types.go` — `LLMMetadataResponse` and all platform metadata structs

## Dependencies

- Depends on: `X_AI_KEY` env var, `github.com/SimonMorphy/grok-go`, X.AI API
- Used by: `main.go`; output fed to `news.ConvertToContentItem` and then the manifest

## Configuration

| Env Var | Description |
|---------|-------------|
| `X_AI_KEY` | X.AI API key for both Grok text and Grok image endpoints |

| Parameter | Value |
|-----------|-------|
| Text model | `grok-3` |
| Image model | `grok-2-image` |
| Text max tokens | 4000 |
| Timeout | 5 minutes (both calls) |
| Image output dir | `backgrounds/` |
