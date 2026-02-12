# TikTok Upload - Usage Guide

## Overview

Your application now supports uploading videos to TikTok using the **Content Posting API** with support for both **production** and **sandbox** environments.

---

## Setup

### 1. Environment Variables

Add to your `.env` file:

**For Production:**
```env
TIKTOK_CLIENT_KEY=your_production_client_key
TIKTOK_CLIENT_SECRET=your_production_client_secret
```

**For Sandbox (Demo/Testing):**
```env
TIKTOK_CLIENT_KEY_SANDBOX=your_sandbox_client_key
TIKTOK_CLIENT_SECRET_SANDBOX=your_sandbox_client_secret
```

---

## Usage Options

### Option 1: Generate and Upload in One Command

**Production:**
```bash
go run main.go -upload-tiktok
```

**Sandbox (for TikTok app review demo):**
```bash
go run main.go -upload-tiktok -sandbox
```

**Upload to Both YouTube and TikTok:**
```bash
go run main.go -upload-youtube -upload-tiktok
```

### Option 2: Upload Existing Videos

**Upload specific video (Production):**
```bash
go run tools/upload_tiktok/main.go news_20251216_222847
```

**Upload specific video (Sandbox):**
```bash
go run tools/upload_tiktok/main.go --sandbox news_20251216_222847
```

**Upload all unposted videos:**
```bash
go run tools/upload_tiktok/main.go --all-unposted
```

**Preview what would be uploaded (dry run):**
```bash
go run tools/upload_tiktok/main.go --dry-run --all-unposted
```

---

## Authentication Flow (First Time Only)

When you run a TikTok upload command for the first time:

> **Note:** Our implementation uses **PKCE (Proof Key for Code Exchange)** for enhanced security, which is required by TikTok's API. This is handled automatically.

1. **Program displays authorization URL**
   ```
   🔐 TikTok OAuth Authorization Required
   ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

   1. Open this URL in your browser:

      https://www.tiktok.com/v2/auth/authorize/?client_key=...&code_challenge=...
   ```

2. **You open the URL in your browser**
   - Sign in to TikTok (if not already)
   - Review permissions requested
   - Click "Authorize"

3. **TikTok redirects** to your callback URL
   - **GitHub Pages option:** Opens `https://slukehart.github.io/content-generation-automation/callback.html`
   - Page displays the authorization code in a nice format with "Copy" button
   - **Loopback IP option:** Redirects to `http://127.0.0.1:8080/callback?code=...`
   - Page may show "This site can't be reached" (this is OK!)
   - Copy the `code` parameter from the page or URL

4. **Paste code back** into terminal
   ```
   Enter authorization code: ABC123XYZ...
   ```

5. **Credentials cached** for future use
   - Saved to `~/.credentials/tiktok-oauth.json`
   - Future uploads won't require browser authorization

---

## How TikTok Upload Works

### Important: Videos Go to Inbox First!

Unlike YouTube (which posts directly), TikTok uploads go to your **inbox for review**:

```
Upload Flow:
1. Our app → TikTok API → Your TikTok inbox ✅
2. You open TikTok app → Review video
3. You add final caption/hashtags
4. You click "Post" → Video goes LIVE on TikTok feed
```

### Why This Matters

- **User Control:** You review every video before it goes public
- **No Spam:** Prevents automated spam/unwanted content
- **Quality Control:** Add final touches in TikTok app
- **TikTok Policy:** API can only upload to inbox, not directly to feed

### After Upload

1. **Check TikTok App Notifications** 📱
2. **Open Video from Inbox**
3. **Add/Edit:**
   - Caption (our app provides suggested caption)
   - Hashtags (our app provides suggested hashtags)
   - Music (optional)
   - Privacy settings
   - Comment/duet/stitch settings
4. **Click "Post"** or **Delete** if you don't want to publish

---

## Sandbox vs Production

### Sandbox Mode (`-sandbox` flag)

**When to Use:**
- Testing your integration
- Creating demo videos for TikTok app review
- Development/debugging

**Limitations:**
- Videos don't go to real TikTok feed
- For testing only
- May have different rate limits

**Usage:**
```bash
go run main.go -upload-tiktok -sandbox
```

### Production Mode (default)

**When to Use:**
- Real video uploads to your actual TikTok account
- Production content distribution

**Usage:**
```bash
go run main.go -upload-tiktok
```

---

## Redirect URI Configuration

Since this is a **command-line/desktop application** (not a web app), TikTok's OAuth requires special handling.

### Option 1: GitHub Pages Callback (Recommended)

**Best for:** TikTok app review and production use

**Setup:**
1. Enable GitHub Pages for this repository
2. Set redirect URI in TikTok Developer Portal:
   ```
   https://slukehart.github.io/content-generation-automation/callback.html
   ```
3. After authorization, you'll see a nice page with the code displayed
4. Click "Copy Code" button and paste into terminal

**Why this works:**
- TikTok requires a public HTTPS URL for non-localhost apps
- Shows TikTok reviewers that your OAuth flow works properly
- Better user experience with visual feedback

### Option 2: Loopback IP Address

**Best for:** Local development if TikTok accepts it

**Setup:**
1. Set redirect URI in TikTok Developer Portal:
   ```
   http://127.0.0.1:8080/callback
   ```
2. Update `tiktokRedirectURI` in `media/tiktok_auth.go` (see comments in code)
3. The redirect may show "site can't be reached" - just copy the code from URL

**Note:** Some OAuth providers reject localhost/loopback IPs. Test this first.

### Option 3: Out-of-Band (OOB)

**Best for:** When nothing else works

**Setup:**
1. Set redirect URI in TikTok Developer Portal:
   ```
   urn:ietf:wg:oauth:2.0:oob
   ```
2. Update `tiktokRedirectURI` in `media/tiktok_auth.go`
3. TikTok shows the code directly on their page (no redirect)

**Current Configuration:** Using GitHub Pages callback (Option 1)

---

## Rate Limits

**TikTok API Rate Limit:** 6 requests per minute per user

**Our Usage:** Typically 1-3 videos per hour (well under limit)

**If you hit the limit:**
- Wait 1 minute and try again
- Use `--dry-run` to preview before uploading
- Upload videos in smaller batches

---

## Troubleshooting

### Error: "TIKTOK_CLIENT_KEY must be set"

**Solution:** Add credentials to `.env` file

```env
TIKTOK_CLIENT_KEY=your_key_here
TIKTOK_CLIENT_SECRET=your_secret_here
```

### Error: "failed to get TikTok access token"

**Solution:** Delete cached credentials and re-authorize

```bash
rm ~/.credentials/tiktok-oauth.json
go run main.go -upload-tiktok
```

### Error: "video file too large"

**Solution:** Video must be < 287.6 MB

- Our videos are typically < 50 MB (no issue)
- Check video file size: `ls -lh news_*.mp4`

### Error: "TikTok API error"

**Solution:** Check TikTok API status

- May be temporary API issues
- Wait a few minutes and retry
- Check if sandbox vs production credentials match your flag usage

### Video not showing in TikTok inbox

**Solution:**
1. Wait 1-2 minutes (processing time)
2. Pull down to refresh in TikTok app
3. Check notification center
4. Verify upload succeeded (check terminal output)

---

## Example Workflows

### Daily News Upload (Production)

```bash
# Morning routine
go run main.go -upload-youtube -upload-tiktok

# Check TikTok app
# Complete posting in app
```

### Demo for TikTok App Review (Sandbox)

```bash
# Generate and upload to sandbox
go run main.go -upload-tiktok -sandbox

# Video shows in sandbox environment
# Screenshot for app review demo video
```

### Batch Upload Existing Videos

```bash
# Preview what will be uploaded
go run tools/upload_tiktok/main.go --dry-run --all-unposted

# Upload all
go run tools/upload_tiktok/main.go --all-unposted

# Check TikTok app for each video
```

---

## Security Notes

### Credentials Storage

- OAuth tokens: `~/.credentials/tiktok-oauth.json`
- **Never commit this file** to git
- User can delete anytime to revoke access

### Scopes Requested

- `user.info.basic` - Get user profile info
- `video.upload` - Upload video files
- `video.publish` - Publish to inbox

### Revoking Access

**Method 1:** Delete local credentials
```bash
rm ~/.credentials/tiktok-oauth.json
```

**Method 2:** Revoke in TikTok app
- Settings → Privacy → Apps & websites
- Find your app → Revoke access

---

## Manifest Tracking

After upload, check manifest:

```bash
go run tools/inspect_manifest.go list
```

**Status Fields:**
- `tiktok.posted`: `false` (uploaded to inbox, not feed)
- `tiktok.url`: `"inbox_uploaded (publish_id: abc123)"`
- `tiktok.posted_at`: Upload timestamp

**After you post in TikTok app:**
- Manually update manifest if tracking public URL needed
- Or leave as-is (inbox upload tracked)

---

## FAQ

**Q: Why can't I post directly to TikTok feed?**
A: TikTok API only supports uploading to inbox. Users must complete posting in the app. This is TikTok's policy to prevent spam.

**Q: Do I need sandbox credentials for production?**
A: No. Sandbox is only for testing/demos. Use production credentials for real uploads.

**Q: Can I upload the same video to both YouTube and TikTok?**
A: Yes! Use both flags: `go run main.go -upload-youtube -upload-tiktok`

**Q: How many videos can I upload per day?**
A: Rate limit is 6 per minute. Daily limit depends on your TikTok account status (typically unlimited for verified creators).

**Q: Does this violate TikTok's terms?**
A: No. We use the official Content Posting API and follow all TikTok guidelines. Videos go to inbox for review, giving users full control.

---

## Support

For issues:
- Check error messages in terminal
- Review troubleshooting section above
- Verify credentials in `.env` file
- Check TikTok API documentation: https://developers.tiktok.com/doc/content-posting-api-reference-upload-video

---

**Version:** 1.0
**Last Updated:** December 17, 2025

