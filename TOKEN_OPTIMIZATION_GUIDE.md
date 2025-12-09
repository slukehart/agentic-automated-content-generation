# Token Optimization Guide for Grok LLM

## Problem Solved
You wanted to process multiple news articles with Grok using a custom prompt, while minimizing token usage by not sending the same prompt repeatedly.

## Solution: Two Strategies Implemented

### Strategy 1: Individual Article Processing with System Message
**Function:** `GenerateNewsReportSummary(article, systemPrompt)`

**How it works:**
- Uses a **system message** for your prompt/instructions
- System messages set the context without being repeated in the conversation
- Each article is processed individually with one API call per article

**Token usage per article:**
```
system_prompt_tokens + article_tokens + response_tokens
```

**When to use:**
- Processing a small number of articles (< 5)
- Need individual error handling per article
- Want to process articles as they come in real-time

---

### Strategy 2: Batch Processing (MOST EFFICIENT) ⭐
**Function:** `GenerateBatchNewsReportSummaries(articles, systemPrompt)`

**How it works:**
- Sends your prompt **ONCE** as a system message
- Combines ALL articles into a single user message
- Makes only **ONE** API call for all articles
- Grok returns all summaries in a single response

**Token usage for all articles:**
```
system_prompt_tokens + (article_1_tokens + article_2_tokens + ... + article_N_tokens) + combined_response_tokens
```

**Savings compared to Strategy 1:**
```
Saves: (N - 1) × system_prompt_tokens
Where N = number of articles
```

**Example savings:**
- 10 articles with a 100-token prompt
- Strategy 1: 10 API calls, 1000 prompt tokens
- Strategy 2: 1 API call, 100 prompt tokens
- **Savings: 900 tokens (90% reduction in prompt tokens!)**

**When to use:**
- Processing multiple articles at once (5+)
- Token efficiency is critical
- Speed matters (1 API call vs N calls)
- Batch processing fits your workflow

---

## Usage Example

```go
package main

import (
    "fmt"
    "log"
    "github.com/slukehart/content-generation-automation/news"
)

func main() {
    // Define your prompt ONCE
    systemPrompt := `You are a news summarization assistant.
    For each article, create a concise 2-3 sentence summary
    focusing on key facts.`

    // Fetch articles
    articles := news.ParseNewsArticles()

    // OPTION 1: Individual processing
    for _, article := range articles {
        summary, err := news.GenerateNewsReportSummary(article, systemPrompt)
        if err != nil {
            log.Printf("Error: %v", err)
            continue
        }
        fmt.Println(summary)
    }

    // OPTION 2: Batch processing (recommended for token efficiency)
    summaries, err := news.GenerateBatchNewsReportSummaries(articles, systemPrompt)
    if err != nil {
        log.Fatalf("Error: %v", err)
    }
    fmt.Println(summaries)
}
```

---

## Key Points

### ✅ What You Get
1. **Prompt sent only once** - No repetition of your instructions
2. **System message usage** - Efficient context setting
3. **Batch processing option** - Maximum token savings
4. **Proper error handling** - Returns errors instead of fatal crashes
5. **Flexible API** - Choose individual or batch based on your needs

### 🔧 Technical Details
- Uses `grok-beta` model
- Temperature: 0.7 (balanced creativity/consistency)
- Max tokens: 500 (individual) / 2000 (batch)
- Context managed via system messages
- Proper Go error handling

### 💰 Cost Savings
For 20 articles with a 150-token prompt:
- **Old approach:** 20 × 150 = 3,000 prompt tokens
- **Strategy 1 (system message):** 20 × 150 = 3,000 prompt tokens (but cleaner architecture)
- **Strategy 2 (batch):** 1 × 150 = 150 prompt tokens
- **Savings: 95% reduction in prompt tokens!**

---

## Recommendations

1. **Use Strategy 2 (Batch)** for most use cases
2. **Use Strategy 1 (Individual)** when:
   - Processing articles one at a time
   - Need granular error handling
   - Real-time processing requirements

3. **Consider adding:**
   - Rate limiting if processing many articles
   - Caching for repeated article processing
   - Retry logic for failed API calls

---

## Environment Setup
Make sure you have in your `.env` file:
```
GROK_API_KEY=your_grok_api_key_here
NEWS_API_KEY=your_news_api_key_here
```

