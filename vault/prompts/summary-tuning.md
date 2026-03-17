# Summary Tuning

## Overview

The news summary is generated as part of the single metadata generation call in `news/metadata_generation.go` → `GenerateEnrichedNewsContent()`. There is no separate summary-only call — the summary is the `summary` field in the JSON response from `MetadataGenerationPrompt()`.

---

## Generation Parameters

| Parameter     | Value         |
|---------------|---------------|
| Model         | grok-3        |
| Temperature   | 0.7           |
| Max tokens    | 4000          |
| Timeout       | 5 minutes     |
| API           | X.AI (x.ai)   |

The 4000-token budget covers the full response: summary + SEO block + all six platform metadata objects. In practice the summary consumes roughly 200–300 tokens, leaving the bulk for platform content.

---

## Target Length and Tone

- **Word count:** 150–200 words
- **Tone:** Neutral, factual, authoritative — broadcast news style
- **Delivery target:** ~60 seconds of spoken audio (average 150 words/min)

The word count and estimated speaking duration are calculated after generation in `ConvertToContentItem()`:

```go
wordCount := len(strings.Fields(enriched.Summary))
estimatedDuration := int(float64(wordCount) / 2.5) // 2.5 words/sec = 150 words/min
```

Both values are stored in `ContentInfo.WordCount` and `ContentInfo.EstimatedDurationSecs` in the manifest.

---

## Prompt-Enforced Summary Rules

From the system prompt:

1. Start with an engaging hook
2. Use neutral, factual, authoritative tone — no speculation, opinions, or emotional language
3. Include all key points, context, background, and verified facts
4. Cite the original source
5. Meaningful conclusion that conveys the article's main point

---

## Quality Observations

**Works well:**
- The broadcast-news framing produces clean, structured output that reads well as a teleprompter script
- The hook requirement gives TTS-generated videos a strong opening line
- 150–200 words fits HeyGen avatar generation constraints — long enough to feel substantive, short enough to not require advanced editing

**Friction points:**
- Temperature 0.7 introduces some variability in word count. Summaries occasionally land below 140 or above 220 words. The manifest stores the actual word count but there is no enforcement or retry if the range is missed.
- The prompt does not specify sentence count or maximum sentence length. Occasionally Grok produces a single run-on sentence as the hook, which creates pacing issues in TTS delivery.
- Grok-3 interprets "cite the original source" inconsistently — sometimes it names the outlet, sometimes it says "according to [URL]", and sometimes it omits attribution entirely.
- For articles behind paywalls, Grok falls back to secondary sources per the fallback instruction. The summary quality is usually acceptable, but the source attribution in the manifest still reflects the original (potentially unread) URL.

**Potential improvements:**
- Add a post-generation word count check and retry if outside 140–210 range
- Lower temperature to 0.4–0.5 for more deterministic output without losing fluency
- Add explicit instruction: "Use 3–5 short sentences. Each sentence should be under 25 words."
- Add source citation format: "According to [outlet name], ..." as required first-sentence attribution format
- Consider a separate summary call with a tighter prompt and lower max_tokens (500–800) to reduce risk of JSON truncation on the metadata side
