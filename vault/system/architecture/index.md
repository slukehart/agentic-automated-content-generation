# Architecture

How each component of the pipeline works.

## Contents

- [[pipeline-overview]] — End-to-end pipeline: news fetch → metadata → video → upload
- [[news-fetching]] — NewsAPI article retrieval, first-article selection
- [[metadata-generation]] — Grok-3 single-call summary + 6-platform metadata generation
- [[video-generation]] — Custom AI avatar video generation (AWS GPU, replacing HeyGen)
- [[manifest-system]] — Thread-safe JSON content manifest with CRUD operations
- [[upload-system]] — YouTube OAuth upload + TikTok PKCE inbox upload
