# Text-to-Speech Video Guide

## Summary Length for 1-Minute Videos

Your summaries are now optimized for text-to-speech video production.

### Speaking Rate Math

**Average TTS Speaking Rate:** 150-160 words per minute

**Target for 1-minute video:**
- Minimum: 150 words
- Target: 150-200 words
- Maximum: ~180 words (for clear, unhurried speech)

### Current Configuration

✅ **Prompt specifies:** "150-200 words minimum per summary"

✅ **MaxTokens set to:** 8,000 tokens
- Supports ~6,000 words of output
- Enough for 30+ detailed summaries
- Or 10 very comprehensive summaries

✅ **Token-to-word ratio:** ~1.33 tokens per word
- 200 words ≈ 266 tokens per summary
- 10 summaries ≈ 2,660 tokens
- Buffer included for safety

## Expected Output Per Article

### Structure
Each summary will be **150-200 words** and include:
1. **Hook** (20-30 words) - Engaging opening
2. **Main Story** (80-120 words) - Key facts, context, background
3. **Source & Wrap** (20-30 words) - Attribution and conclusion

### Example Timeline
For a 1-minute video per article:
- 0:00-0:10 - Hook (engaging opening)
- 0:10-0:50 - Main content (detailed facts)
- 0:50-1:00 - Source attribution and conclusion

## TTS Recommendations

### Best TTS Services for News
1. **ElevenLabs** - Most natural, news anchor quality
2. **Google Cloud TTS** - Good for News/Broadcast voices
3. **Amazon Polly** - Joanna/Matthew voices (news style)
4. **Azure TTS** - Jenny/Guy Neural voices

### Voice Settings
- **Speed:** 1.0x (normal) or 0.9x (more authoritative)
- **Style:** News/Broadcast/Authoritative
- **Pauses:** Enable for punctuation

## Testing Your Output

### Word Count Check
```bash
# After running your Go program, check word count
echo "Summary text here" | wc -w
```

### Expected Results
- 10 articles = 10 summaries
- Each 150-200 words
- Total: 1,500-2,000 words
- Total speaking time: 10-13 minutes

### Quality Checklist
- [ ] Each summary is 150-200 words
- [ ] Includes engaging hook
- [ ] Contains all key facts
- [ ] Has proper context/background
- [ ] Cites original source
- [ ] Neutral, factual tone
- [ ] No speculation or opinions

## Token Usage

### Per Batch (10 articles)
```
System prompt:     ~180 tokens
Article titles:    ~500 tokens
Response (10×200): ~2,660 tokens
Total:             ~3,340 tokens per batch
```

### Cost Estimation (Grok pricing)
- Input: ~680 tokens × $5/1M = $0.0034
- Output: ~2,660 tokens × $15/1M = $0.0399
- **Total per batch: ~$0.043**

For 100 articles (10 batches):
- **Total cost: ~$0.43**
- **Total video time: ~100 minutes**

## Video Production Workflow

1. **Run the script**
   ```bash
   go run main.go
   ```

2. **Get summaries** (150-200 words each)

3. **Convert to speech** using TTS service

4. **Expected output:**
   - 10 audio files
   - Each ~60 seconds long
   - Ready for video editing

5. **Add visuals:**
   - News graphics
   - Article images
   - Lower thirds with source

## Adjusting Length

### If summaries are too short (< 150 words):
Edit `main.go` prompt:
```go
- Each summary must be 150-200 words minimum
+ Each summary must be 200-250 words minimum
```

And increase MaxTokens:
```go
MaxTokens: 10000, // More room for longer summaries
```

### If summaries are too long (> 200 words):
Edit `main.go` prompt:
```go
- Each summary must be 150-200 words minimum
+ Each summary must be exactly 150-180 words
```

## Pro Tips

1. **Test one article first** - Make sure length is right
2. **Use word count verification** - Check actual output
3. **Adjust temperature** - Lower (0.5) = more consistent length
4. **Monitor token usage** - Stay within budget
5. **Batch processing** - Most cost-effective approach

---

**Current Status:** ✅ Optimized for 1-minute TTS videos
**Word Target:** 150-200 words per summary
**Token Efficiency:** Maximum (batch processing with system message)

