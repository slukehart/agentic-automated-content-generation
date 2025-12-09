# Audio + Video Setup Guide (Google Cloud TTS)

## Overview

This system creates **narrated news videos** using:
1. **Google Cloud Text-to-Speech** - High-quality audio narration
2. **FAL AI** - Video generation
3. **ffmpeg** - Merging audio and video

## Why Google Cloud TTS?

- ✅ **Best for GCP deployment** - Native integration with Google Cloud
- ✅ **Cost-effective** - $4 per 1 million characters
- ✅ **High quality** - Neural voices sound natural
- ✅ **Simple setup** - One service to manage
- ✅ **Scalable** - Built for production workloads

## Prerequisites

### 1. Install ffmpeg

```bash
# macOS
brew install ffmpeg

# Ubuntu/Debian
sudo apt update && sudo apt install ffmpeg

# Verify
ffmpeg -version
```

### 2. Install Python Dependencies

```bash
poetry install
```

### 3. Set Up Google Cloud TTS

#### Step 1: Enable the API

1. Go to [Google Cloud Console](https://console.cloud.google.com/)
2. Create a project (or select existing)
3. Enable **Cloud Text-to-Speech API**:
   - Go to APIs & Services → Library
   - Search for "Text-to-Speech"
   - Click "Enable"

#### Step 2: Create Service Account

1. Go to **IAM & Admin** → **Service Accounts**
2. Click "Create Service Account"
3. Name: `tts-service-account`
4. Grant role: **Cloud Text-to-Speech User**
5. Click "Done"

#### Step 3: Create and Download Key

1. Click on your new service account
2. Go to **Keys** tab
3. Click "Add Key" → "Create new key"
4. Choose **JSON** format
5. Save the file (e.g., `google-tts-credentials.json`)

#### Step 4: Set Environment Variable

Add to your `.env` file:

```bash
GOOGLE_APPLICATION_CREDENTIALS=/absolute/path/to/your/google-tts-credentials.json
```

**Important:** Use the absolute path to the JSON file!

### 4. Set Other API Keys

In your `.env` file:

```bash
# Video generation
FAL_KEY=your-fal-api-key

# News
NEWS_API_KEY=your-newsapi-key
GROK_API_KEY=your-grok-key

# Google Cloud TTS
GOOGLE_APPLICATION_CREDENTIALS=/path/to/google-tts-credentials.json
```

## Usage

### Run Complete Workflow

```bash
go run main.go
```

This will:
1. Fetch news articles
2. Generate AI summaries
3. Create audio narration (Google Cloud TTS)
4. Generate videos (FAL AI)
5. Merge audio + video
6. Output final narrated videos

### Output Files

```
news_1_final.mp4  ← Final video with narration
news_2_final.mp4
news_3_final.mp4
...
```

## Voice Customization

### Available Voices

Google Cloud TTS offers many professional voices. Edit `audio/tts.go` line 58:

```go
func GenerateNewsAudio(newsSummary string, outputPath string) (*TTSResponse, error) {
    return GenerateAudioWithOptions(
        newsSummary,
        outputPath,
        "en-US-Neural2-J", // Change this voice
        1.0,               // Adjust speed if needed
    )
}
```

### Popular News Voices:

**Male:**
- `en-US-Neural2-J` - Professional, authoritative (default)
- `en-US-Neural2-D` - Warm, conversational
- `en-GB-Neural2-B` - British accent, formal

**Female:**
- `en-US-Neural2-F` - Professional, clear
- `en-US-Neural2-C` - Friendly, engaging
- `en-GB-Neural2-A` - British accent, elegant

**See all voices:** https://cloud.google.com/text-to-speech/docs/voices

### Adjust Speaking Speed

```go
return GenerateAudioWithOptions(
    newsSummary,
    outputPath,
    "en-US-Neural2-J",
    1.2,  // 1.2x faster (range: 0.5 to 2.0)
)
```

## Testing

### Test Google Cloud TTS

```bash
# Test Python script directly
echo '{"text":"This is a test","output_path":"test.mp3"}' | \
    poetry run python audio/tts_generation.py

# Should create test.mp3
```

### Test Audio + Video Merge

```bash
# Generate test audio
poetry run python audio/tts_generation.py "Test narration" test_audio.mp3

# Generate test video
poetry run python video/video_generation.py text "Test video" test_video.mp4

# Merge them
ffmpeg -i test_video.mp4 -i test_audio.mp3 -c:v copy -c:a aac -shortest test_final.mp4
```

## Cost Estimation

### Google Cloud TTS Pricing

- **Standard voices:** $4 per 1M characters
- **Neural voices:** $16 per 1M characters (we use these)
- **WaveNet voices:** $16 per 1M characters

### Example Calculation (10 news articles):

- 200 words per article = 1,000 characters
- 10 articles = 10,000 characters
- Neural TTS cost: **$0.16**

### Plus Video Generation:

- 10 videos @ $0.05-0.10 each
- Video cost: **$0.50-1.00**

**Total per run:** ~$0.66-1.16

### Monthly Cost (daily runs):

- 30 runs × $0.91 average = **~$27/month**

**Much cheaper than other premium TTS services!**

## Troubleshooting

### "Could not load credentials"

Make sure:
1. `GOOGLE_APPLICATION_CREDENTIALS` path is absolute (not relative)
2. The JSON file exists and is readable
3. The service account has "Text-to-Speech User" role

```bash
# Check if file exists
ls -la $GOOGLE_APPLICATION_CREDENTIALS

# Test credentials
gcloud auth activate-service-account --key-file=$GOOGLE_APPLICATION_CREDENTIALS
```

### "API not enabled"

1. Go to https://console.cloud.google.com/apis/library
2. Search "Text-to-Speech"
3. Click "Enable"
4. Wait a few minutes for propagation

### "ffmpeg not found"

```bash
# Install ffmpeg
brew install ffmpeg  # macOS
sudo apt install ffmpeg  # Linux
```

### Audio/Video length mismatch

If audio is longer than video (5 seconds), you'll hear cut-off narration.

**Solution 1:** Increase video duration

```go
// In video/video.go line 40
Duration:   10,  // Increase from 5 to 10 seconds
```

**Solution 2:** Speed up speech

```go
// In audio/tts.go line 61
1.5,  // 1.5x faster speech
```

### Poor audio quality

Google Cloud Neural voices should sound very natural. If not:
1. Check you're using Neural2 voices (not Standard)
2. Try different voices
3. Verify the service account has proper permissions

## Google Cloud Deployment

### For Cloud Run / Cloud Functions:

Your code is already set up! Just:

1. Use Secret Manager for credentials:
```go
// Instead of .env file
credentials := getSecretFromGCP("google-tts-credentials")
```

2. Mount credentials in container or use Workload Identity

3. The code automatically uses `GOOGLE_APPLICATION_CREDENTIALS`

### Workload Identity (Recommended for production):

Instead of a JSON key file, use Workload Identity:

```bash
# No JSON file needed!
# Just grant the compute service account TTS permissions
gcloud projects add-iam-policy-binding PROJECT_ID \
  --member="serviceAccount:PROJECT_NUMBER-compute@developer.gserviceaccount.com" \
  --role="roles/cloudtexttospeech.user"
```

Then remove `GOOGLE_APPLICATION_CREDENTIALS` from your .env - it will use the default compute credentials automatically!

## Advanced Usage

### Custom Voice Parameters

```go
// Adjust pitch, volume, etc. by modifying Python script
// In audio/tts_generation.py, modify audio_config:

audio_config = texttospeech.AudioConfig(
    audio_encoding=texttospeech.AudioEncoding.MP3,
    speaking_rate=speed,
    pitch=0.0,          # -20 to 20
    volume_gain_db=0.0, # Volume adjustment
)
```

### Multiple Languages

```go
// In audio/tts.go, use different voice:
"es-ES-Neural2-A",  // Spanish
"fr-FR-Neural2-B",  // French
"de-DE-Neural2-C",  // German
```

### SSML Support

For more control over pronunciation, use SSML in your text:

```text
<speak>
  Breaking news: The stock market <emphasis level="strong">surged</emphasis> today.
  <break time="500ms"/>
  Details at 11.
</speak>
```

## Next Steps

1. ✅ Test the audio + video workflow
2. Try different voices and speeds
3. Adjust video duration to match audio length
4. Deploy to Google Cloud with Workload Identity
5. Add background music or sound effects

---

**Questions?** Check the main README.md or consult [Google Cloud TTS Docs](https://cloud.google.com/text-to-speech/docs)

