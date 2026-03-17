# TikTok Integration Workflow Diagram

## Overview
This document illustrates how our AI-powered news automation application integrates with **TikTok's Content Posting API** (Upload to TikTok) using TikTok's OAuth 2.0 authentication system.

## TikTok Products & APIs Used

### 1. Login Kit - OAuth 2.0 Authentication
**Purpose:** Secure user authentication and authorization
**Documentation:** https://developers.tiktok.com/doc/login-kit-overview
**What we use:**
- OAuth 2.0 authorization flow
- Token management (access & refresh tokens)
- User authentication

**Scopes Requested:**
- `user.info.basic` - Get user profile information
- `video.upload` - Upload video files to TikTok
- `video.publish` - Publish videos to user's inbox

### 2. Content Posting API - Upload to TikTok
**Purpose:** Upload video drafts to creator's TikTok inbox
**Product Page:** https://developers.tiktok.com/products/content-posting-api/
**API Reference:** https://developers.tiktok.com/doc/content-posting-api-reference-upload-video
**What we use:**
- Initialize upload session
- Upload video file (MP4 format)
- Publish to user's inbox for review

**Why "Upload to TikTok" (not Direct Post):**
- Gives creators control to review content before posting
- Allows adding final captions, hashtags, and settings in TikTok app
- Prevents automated spam or unwanted content
- Aligns with TikTok's quality and authenticity guidelines

---

## Complete Workflow Diagram

```
┌─────────────────────────────────────────────────────────────────────┐
│                    AUTOMATED NEWS VIDEO WORKFLOW                     │
│                   (TikTok Content Posting API Integration)           │
└─────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 1: CONTENT GENERATION (No User Data - Public News Only)      │
└──────────────────────────────────────────────────────────────────────┘

    ┌─────────────┐
    │  NewsAPI    │  ← Fetch verified public news articles
    │  (Public)   │
    └──────┬──────┘
           │
           ▼
    ┌─────────────┐
    │  Grok AI    │  ← Summarize article (150-200 words)
    │ Summarizer  │     • NO user data involved
    └──────┬──────┘     • Only public news text
           │
           ▼
    ┌─────────────┐
    │   HeyGen    │  ← Generate AI avatar video (60 seconds)
    │ AI Avatars  │     • Professional news presenter
    └──────┬──────┘     • Portrait 9:16 format (TikTok-optimized)
           │
           ▼
    ┌──────────────────────────┐
    │  news_YYYYMMDD_final.mp4 │  ← Ready for distribution
    │  • Duration: ~60 seconds  │
    │  • Format: MP4, 9:16      │
    │  • Resolution: 1080x1920  │
    │  • Source attribution     │
    │  • AI disclosure          │
    └────────────┬─────────────┘
                 │
                 │
┌────────────────┴────────────────────────────────────────────────────┐
│  PHASE 2: TIKTOK AUTHENTICATION (One-Time Setup)                    │
└──────────────────────────────────────────────────────────────────────┘

    ┌────────────────┐
    │  Content Creator│  ← User initiates: "Upload to TikTok"
    │  (User)        │
    └────────┬───────┘
             │
             ▼
    ┌──────────────────────┐
    │  TikTok Login Kit    │  ← Scope: user.info.basic
    │  OAuth 2.0 Flow      │
    └──────────┬───────────┘
               │
               ▼
    ┌────────────────────────────────┐
    │  1. Browser Opens              │  ← System browser launched
    │  2. User Logs into TikTok      │
    │  3. User Authorizes App        │
    │  4. Receives Auth Code         │
    └──────────┬─────────────────────┘
               │
               ▼
    ┌────────────────────────────┐
    │  Token Exchange            │  ← Exchange code for tokens
    │  • Access Token            │
    │  • Refresh Token           │
    └──────────┬─────────────────┘
               │
               ▼
    ┌────────────────────────────────┐
    │  Store Locally                 │  ← ~/.credentials/tiktok-oauth.json
    │  • Tokens cached on user device│     • NOT on our servers
    │  • Used for future uploads     │     • User controls deletion
    └────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 3: VIDEO UPLOAD TO TIKTOK                                    │
│  Using: TikTok Content Posting API - Upload to TikTok               │
│  Reference: https://developers.tiktok.com/products/content-posting-api/ │
└──────────────────────────────────────────────────────────────────────┘

    ┌────────────────┐
    │  Our App CLI   │  ← Command: go run main.go -upload-tiktok
    └────────┬───────┘
             │
             ▼
    ┌──────────────────────────┐
    │  Read Cached OAuth Tokens│  ← Load from ~/.credentials/tiktok-oauth.json
    │  (From Login Kit)        │
    └──────────┬───────────────┘
               │
               ▼
    ┌─────────────────────────────────────────────────────────────┐
    │  Content Posting API - Upload Flow                          │
    │  Endpoint: https://open.tiktokapis.com/v2/post/publish/     │
    │  Documentation: https://developers.tiktok.com/doc/           │
    │                content-posting-api-reference-upload-video    │
    └──────────┬──────────────────────────────────────────────────┘
               │
               ▼
    ┌─────────────────────────────────┐
    │  Step 1: Initialize Upload      │  ← POST /v2/post/publish/inbox/video/init/
    │  • Send: video size, chunk info │     Required Scope: video.upload
    │  • Receive: upload_url          │     Authorization: Bearer {access_token}
    │  • Receive: publish_id           │
    └──────────┬──────────────────────┘
               │
               ▼
    ┌─────────────────────────────────┐
    │  Step 2: Upload Video File      │  ← PUT to upload_url (from Step 1)
    │  • Send: MP4 video binary data  │     Required Scope: video.upload
    │  • Chunked if >5MB              │     Content-Type: video/mp4
    │  • Max size: 287.6 MB           │
    └──────────┬──────────────────────┘
               │
               ▼
    ┌─────────────────────────────────┐
    │  Step 3: Publish to Inbox       │  ← Uses publish_id from Step 1
    │  • Video → User's TikTok Inbox  │     Required Scope: video.publish
    │  • Status: "inbox_uploaded"     │     Video ready for user review
    │  • User receives notification   │
    └──────────┬──────────────────────┘
               │
               ▼
    ┌────────────────────────────────┐
    │  ✅ SUCCESS                     │
    │  Video uploaded to creator's   │
    │  TikTok inbox for review       │
    └────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────┐
│  PHASE 4: USER COMPLETES POST IN TIKTOK APP (User Control)          │
└──────────────────────────────────────────────────────────────────────┘

    ┌────────────────────────────┐
    │  Creator Opens TikTok App  │  ← Notifications alert user
    └────────────┬───────────────┘
                 │
                 ▼
    ┌────────────────────────────────┐
    │  Reviews Video in Inbox        │  ← User sees uploaded video
    │  • Plays/previews video        │
    │  • Checks content quality      │
    └────────────┬───────────────────┘
                 │
                 ▼
    ┌────────────────────────────────────┐
    │  Edits & Adds Metadata             │  ← User has FULL control
    │  • Write/edit caption              │
    │  • Add hashtags (#news #shorts)    │
    │  • Select music (optional)         │
    │  • Set privacy (public/private)    │
    │  • Enable comments/duets           │
    │  • Add location tag (optional)     │
    └────────────┬───────────────────────┘
                 │
                 ▼
    ┌────────────────────────────┐
    │  DECISION POINT            │
    │  Post or Delete?           │
    └─────────┬──────────────────┘
              │
         ┌────┴────┐
         ▼         ▼
    ┌────────┐  ┌────────┐
    │ DELETE │  │  POST  │  ← User clicks "Post"
    └────────┘  └────┬───┘
                     │
                     ▼
           ┌──────────────────┐
           │  Published to    │  ← Video goes LIVE on TikTok
           │  TikTok Feed!    │     • Public sees video
           └──────────────────┘     • Appears in For You Page


┌──────────────────────────────────────────────────────────────────────┐
│  DATA FLOW & PRIVACY                                                 │
└──────────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────────┐
│                    WHAT GOES WHERE                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  TO AI SERVICES (Grok, HeyGen):                                     │
│  ✅ Public news articles (from internet)                            │
│  ❌ NO user data                                                     │
│  ❌ NO personal information                                          │
│  ❌ NO viewing preferences                                           │
│                                                                      │
│  TO TIKTOK API:                                                      │
│  ✅ OAuth tokens (for authentication)                               │
│  ✅ Video file (MP4)                                                 │
│  ❌ NO browsing history                                              │
│  ❌ NO analytics data                                                │
│                                                                      │
│  STORED LOCALLY (User's Device):                                    │
│  ✅ OAuth credentials (~/.credentials/)                             │
│  ✅ Video files (local directory)                                   │
│  ✅ Content metadata (content_manifest.json)                        │
│                                                                      │
│  STORED ON OUR SERVERS:                                             │
│  ❌ NOTHING - We don't have servers!                                │
│  ❌ All data stays on user's device                                 │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────┐
│  SCOPES USAGE BREAKDOWN                                              │
└──────────────────────────────────────────────────────────────────────┘

┌─────────────────────┬──────────────────────────────────────────────┐
│  SCOPE              │  HOW WE USE IT                               │
├─────────────────────┼──────────────────────────────────────────────┤
│  user.info.basic    │  • Get user's TikTok username                │
│                     │  • Display in our CLI for confirmation       │
│                     │  • Verify correct account connected          │
├─────────────────────┼──────────────────────────────────────────────┤
│  video.upload       │  • Upload video file to TikTok servers       │
│                     │  • Initialize upload session                 │
│                     │  • Transfer MP4 data                         │
├─────────────────────┼──────────────────────────────────────────────┤
│  video.publish      │  • Publish video to user's inbox            │
│                     │  • Create draft for user review              │
│                     │  • Does NOT auto-post to feed                │
└─────────────────────┴──────────────────────────────────────────────┘


┌──────────────────────────────────────────────────────────────────────┐
│  RATE LIMITS & COMPLIANCE                                            │
└──────────────────────────────────────────────────────────────────────┘

Rate Limit: 6 requests per minute (per TikTok guidelines)
Our Usage: Typically 1-3 videos per hour (well under limit)

Compliance Measures:
✅ All videos include source attribution
✅ Clear AI disclosure in video description
✅ No copyright infringement (public news summaries)
✅ No misleading content
✅ Adheres to TikTok Community Guidelines
✅ User has final editorial control


┌──────────────────────────────────────────────────────────────────────┐
│  EXAMPLE USE CASE                                                    │
└──────────────────────────────────────────────────────────────────────┘

📰 Morning News Routine:

1. 8:00 AM - Content creator runs: `go run main.go -upload-tiktok`
2. 8:01 AM - App fetches breaking news from NewsAPI
3. 8:02 AM - AI generates 60-second news summary video
4. 8:03 AM - Video uploads to creator's TikTok inbox
5. 8:05 AM - Creator receives TikTok notification
6. 8:10 AM - Creator reviews video, adds caption:
            "🚨 Breaking: [Headline] #news #breakingnews"
7. 8:12 AM - Creator clicks "Post" → Video goes LIVE
8. 8:15 AM - Video appears in followers' feeds

Result: Fresh, timely news content with 12 minutes of automation
(vs. hours of manual video editing)


┌──────────────────────────────────────────────────────────────────────┐
│  TECHNOLOGY STACK                                                    │
└──────────────────────────────────────────────────────────────────────┘

Backend:        Go (Golang)
Video AI:       HeyGen API (AI avatars)
Text AI:        Grok AI (summarization)
News Source:    NewsAPI (verified publishers)
TikTok:         Content Posting API v2, Login Kit
Storage:        Local filesystem (no cloud)
Auth:           OAuth 2.0 with token caching


┌──────────────────────────────────────────────────────────────────────┐
│  SECURITY & PRIVACY HIGHLIGHTS                                       │
└──────────────────────────────────────────────────────────────────────┘

🔒 OAuth tokens stored locally (not on servers)
🔒 No centralized database of user data
🔒 No tracking or analytics beyond basic usage
🔒 Open-source codebase (transparency)
🔒 User can revoke access anytime via TikTok settings
🔒 User controls what gets posted (inbox review)
🔒 GDPR & CCPA compliant privacy policy


┌──────────────────────────────────────────────────────────────────────┐
│  BENEFITS FOR TIKTOK ECOSYSTEM                                       │
└──────────────────────────────────────────────────────────────────────┘

✅ High-quality, factual news content on platform
✅ Helps creators maintain consistent posting schedule
✅ Reduces barrier to entry for news content creation
✅ Promotes transparency with AI disclosure
✅ Proper attribution drives traffic to news sources
✅ No spam or low-quality content (user reviews each video)
✅ Aligns with TikTok's mission of authentic creativity


┌──────────────────────────────────────────────────────────────────────┐
│  CONTACT & SUPPORT                                                   │
└──────────────────────────────────────────────────────────────────────┘

Developer: lukehartsc@gmail.com
GitHub: https://github.com/slukehart/agentic-automated-content-generation
Privacy Policy: https://slukehart.github.io/agentic-automated-content-generation/PRIVACY_POLICY
Terms of Service: https://slukehart.github.io/agentic-automated-content-generation/TERMS_OF_SERVICE
```

---

## Key Takeaways for Review Board

1. **User Control**: Users review and approve every video before it goes public
2. **Privacy First**: No user data used for AI content generation
3. **Transparency**: Clear AI disclosure and source attribution
4. **Platform Benefit**: High-quality news content that follows guidelines
5. **Responsible Use**: Rate limits respected, no spam, proper moderation

---

**Last Updated**: December 17, 2025
**Version**: 1.0

