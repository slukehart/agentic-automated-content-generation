---
type: architecture
component: news-fetching
status: current
created: 2026-03-17
updated: 2026-03-17
tags: [news, newsapi]
---

## Overview

The news-fetching component retrieves the top US headline from NewsAPI and returns the article URL and title for downstream processing. It is intentionally minimal: article selection is deterministic (always first result), so there is no filtering, deduplication, or ranking logic here.

## How It Works

`ParseNewsArticles()` in `news/parseNewsArticles.go`:

1. Loads `.env` via `godotenv.Load()` to get `NEWS_API_KEY`
2. Makes an HTTP GET to `https://newsapi.org/v2/top-headlines?country=us&sortBy=popularity&apiKey={key}`
3. Deserializes the JSON response into `NewsAPIResponse{Status, TotalResults, Articles[]}`
4. Iterates all articles and collects them as `[]AiArticleParameters`
5. Returns `newsReports[0]` — always the first article in the list

The function calls `log.Fatal` on any error (network, JSON parse, .env load), so failures halt the pipeline immediately.

## Data Structures

```go
type AiArticleParameters struct {
    ArticleUrl   string `json:"articleUrl"`
    ArticleTitle string `json:"articleTitle"`
}

type Article struct {
    Source      Source `json:"source"`
    Title       string `json:"title"`
    Author      string `json:"author"`
    Description string `json:"description"`
    Url         string `json:"url"`
    UrlToImage  string `json:"urlToImage"`
    PublishedAt string `json:"publishedAt"`
    Content     string `json:"content"`
}
```

Note: `Article` has more fields than `AiArticleParameters`. Only `Url` and `Title` are passed forward; author, description, image, and content are discarded at this stage.

## Key Files

- `news/parseNewsArticles.go` — contains `ParseNewsArticles()`, all type definitions, and the NewsAPI HTTP call

## Dependencies

- Depends on: `NEWS_API_KEY` env var, NewsAPI v2 endpoint
- Used by: `main.go` (result passed to `news.GenerateEnrichedNewsContent`)

## Configuration

| Env Var | Description |
|---------|-------------|
| `NEWS_API_KEY` | NewsAPI authentication key |

**Endpoint:** `GET https://newsapi.org/v2/top-headlines?country=us&sortBy=popularity`

**Article selection:** first element of the returned articles array (index 0). No randomization or history-based deduplication.
