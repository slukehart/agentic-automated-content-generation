---
type: integration
service: newsapi
category: infrastructure
status: implemented
auth_method: api-key
created: 2026-03-17
updated: 2026-03-17
tags: [news, api, fetching]
---

## Overview

Fetch the top US news headline to use as the source article for each content generation run. NewsAPI provides the article title and URL that are passed downstream to Grok for summarization and metadata generation.

## API Details

- **Endpoint**: `GET https://newsapi.org/v2/top-headlines`
- **Query parameters**:
  - `country=us`
  - `sortBy=popularity`
  - `apiKey={NEWS_API_KEY}`
- **Auth**: API key appended as query parameter `apiKey`
- **Key env var**: `NEWS_API_KEY` (loaded from `.env` via `godotenv`)

Source: `news/parseNewsArticles.go` — `ParseNewsArticles()`.

## Request

```
GET https://newsapi.org/v2/top-headlines?country=us&sortBy=popularity&apiKey={NEWS_API_KEY}
```

No request body. Standard HTTP GET.

## Response

Returns a `NewsAPIResponse` with:
- `status` — `"ok"` on success
- `totalResults` — total number of articles
- `articles` — array of `Article` objects

Each `Article` contains: `source` (id, name), `title`, `author`, `description`, `url`, `urlToImage`, `publishedAt`, `content`.

## Output

`ParseNewsArticles()` maps the articles array to `[]AiArticleParameters` and returns `newsReports[0]` — the first (most popular) article.

`AiArticleParameters`:
```go
type AiArticleParameters struct {
    ArticleUrl   string
    ArticleTitle string
}
```

## Notes

- Only the first article is used per pipeline run. The full list is built but discarded after index 0.
- `log.Fatal` is called on any error (network, JSON parse, .env load). No graceful error handling.
- `NewsAPI` free tier limits requests and does not return full article body — only `description` and a truncated `content` field. The downstream Grok call uses the URL for full context.
