---
type: architecture
component: manifest
status: current
created: 2026-03-17
updated: 2026-03-17
tags: [manifest, data, storage]
---

## Overview

The manifest system is the persistent data store for all generated content. It is a single JSON file (`content_manifest.json`) managed by `ManifestManager` in the `metadata` package. It tracks every piece of content from generation through upload and provides the source of truth for posting status across all platforms.

## How It Works

`ManifestManager` wraps the JSON file with a `sync.RWMutex` for thread-safe reads and writes. All operations load the file from disk, modify in memory, and write the full file back — there is no partial update or append-only write at the byte level, but semantically it is append-only (items are added, never deleted in normal operation).

### Operations

| Method | Lock | Description |
|--------|------|-------------|
| `CreateManifest()` | Write | Creates an empty manifest file |
| `LoadManifest()` | Read | Reads and parses the file; returns empty manifest if file doesn't exist |
| `SaveManifest(manifest)` | Write | Marshals and writes the full manifest |
| `AddItem(item)` | Write (via Save) | Appends a new item; errors on duplicate ID |
| `UpdateItem(item)` | Write (via Save) | Replaces an existing item by ID; errors if not found |
| `GetItem(id)` | Read (via Load) | Returns a single item by ID |
| `GetAllItems()` | Read (via Load) | Returns all items |
| `GetItemsByStatus(platform, posted)` | Read (via Load) | Filters items by posting status for a given platform |
| `DeleteItem(id)` | Write (via Save) | Removes an item by ID |
| `GetStats()` | Read (via Load) | Returns total item count and per-platform posted counts |

`saveManifestUnsafe` always updates `generated_at` to `time.Now()` before writing.

## Schema

```go
type ContentManifest struct {
    Version     string        `json:"version"`      // Currently "1.0"
    GeneratedAt time.Time     `json:"generated_at"`
    Items       []ContentItem `json:"items"`
}
```

### ContentItem Fields

Each `ContentItem` contains:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique ID, format: `news_YYYYMMDD_HHMMSS` |
| `created_at` | time.Time | When the item was generated |
| `source` | `SourceInfo` | Original article URL, title, author, published date, source domain |
| `content` | `ContentInfo` | Summary text, word count, estimated duration in seconds |
| `media` | `MediaInfo` | Audio/video file paths, duration, resolution (`"1280x720"`), avatar ID |
| `seo` | `SEOInfo` | Primary keywords, secondary keywords, topics, sentiment, target audience |
| `platforms` | `PlatformMetadata` | Platform-specific metadata for 6 platforms (see below) |
| `posting_status` | `PostingStatus` | Per-platform `{posted bool, url, posted_at, error}` |
| `analytics` | `AnalyticsInfo` | Views/engagement maps (future use, omitted if empty) |

### Platform Metadata

`PlatformMetadata` has a field for each of 6 platforms:
- `youtube` — title, description, tags, category_id, default_language, privacy_status, timestamps
- `tiktok` — caption, hashtags, privacy_level, duet_enabled, stitch_enabled
- `instagram` — caption, hashtags, location, collaborators
- `twitter` — tweet, hashtags, reply_settings
- `facebook` — message, link_description
- `linkedin` — post_text, hashtags

### Posting Status

Each platform entry in `PostingStatus` uses `PlatformStatus`:

```go
type PlatformStatus struct {
    Posted   bool       `json:"posted"`
    URL      *string    `json:"url,omitempty"`
    PostedAt *time.Time `json:"posted_at,omitempty"`
    Error    *string    `json:"error,omitempty"`
}
```

For TikTok inbox uploads, `Posted` remains `false` (not yet live on feed) and `URL` stores `"inbox_uploaded (publish_id: {id})"`.

## Key Files

- `metadata/manifest.go` — `ManifestManager`, all CRUD methods
- `metadata/types.go` — `ContentManifest`, `ContentItem`, and all nested struct definitions

## Dependencies

- Depends on: local filesystem, `encoding/json`, `sync`
- Used by: `main.go` (AddItem after generation, UpdateItem after each upload attempt)

## Configuration

| Constant | Value | Description |
|----------|-------|-------------|
| `ManifestFileName` | `"content_manifest.json"` | Default file path (project root) |
| `ManifestVersion` | `"1.0"` | Manifest format version |

The manifest file is created automatically on first `AddItem` call if it does not exist. It is stored in the project root alongside the binary.
