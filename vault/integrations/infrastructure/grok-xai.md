---
type: integration
service: grok-xai
category: infrastructure
status: implemented
auth_method: api-key
created: 2026-03-17
updated: 2026-03-17
tags: [grok, xai, llm, images]
---

## Overview

Two X.AI APIs are used in the pipeline:

1. **grok-3 chat completions** — Generate the news summary and all platform metadata (titles, captions, hashtags, tags) in a single API call.
2. **grok-2-image image generation** — Generate a photorealistic newsroom background image for use in video composition.

## Authentication

- **Auth**: Bearer token via `Authorization: Bearer {X_AI_KEY}` header
- **Key env var**: `X_AI_KEY` (loaded from `.env` via `godotenv`)
- **Go client**: `github.com/SimonMorphy/grok-go`

## API 1: grok-3 Chat Completions

**Purpose**: Generate news summary + full platform metadata in one call.

**Client setup**:
```go
client, err := grok.NewClientWithOptions(apiKey, grok.WithTimeout(5*time.Minute))
```

**Request parameters**:
- `Model`: `"grok-3"`
- `Temperature`: `0.7`
- `MaxTokens`: `4000` (summary + all platform metadata)
- `StreamOptions.IncludeUsage`: `true`
- Context timeout: 5 minutes

**Usage**:
- `GenerateEnrichedNewsContent()` — primary function; sends system prompt (from `metadata.MetadataGenerationPrompt()`) and user prompt (from `metadata.UserPromptForArticle(title, url)`); returns `EnrichedNewsContent` with `Summary` and `LLMMetadataResponse`.
- `GenerateBatchNewsReportSummaries()` — deprecated batch function; summary-only, max tokens 8000.

**Response handling**: The LLM returns JSON (sometimes wrapped in markdown code fences). The code strips ` ```json ` / ` ``` ` before unmarshaling into `metadata.LLMMetadataResponse`.

Source: `news/parseNewsArticles.go`, `news/metadata_generation.go`.

## API 2: grok-2-image Image Generation

**Purpose**: Generate a newsroom background image for video composition.

**Endpoint**: `POST https://api.x.ai/v1/images/generations`

**Request body**:
```json
{
  "model": "grok-2-image",
  "prompt": "...",
  "n": 1,
  "response_format": "b64_json"
}
```

**Response**: `data[0].b64_json` — base64-encoded JPEG image (Grok returns JPEG regardless of format requested).

**Client**: Raw `net/http` with 5-minute timeout (not the grok-go client).

**Output**: Image decoded from base64 and saved to `backgrounds/newsroom_{timestamp}.jpg`.

**Default prompt** (used when none provided):
> "A professional modern newsroom background with dark blue tones, large screens displaying news graphics, sleek furniture, and ambient lighting. Cinematic, high quality, photorealistic."

Source: `news/metadata_generation.go` — `GenerateNewsroomBackground()`.

## Notes

- Both APIs share the same `X_AI_KEY`.
- The chat completions call is the most latency-sensitive step — budget 1–2 minutes per run.
- If the LLM wraps its JSON in markdown code blocks, the stripping logic in `GenerateEnrichedNewsContent()` handles it.
- Image generation is a separate, optional step from metadata generation and can be called independently.

## Related

- [[metadata-generation]] — How the pipeline uses Grok-3 for metadata
- [[summary-tuning]] — Quality tuning for Grok summary output
