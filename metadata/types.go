package metadata

import "time"

// ContentManifest represents the entire collection of generated content items
type ContentManifest struct {
	Version     string        `json:"version"`
	GeneratedAt time.Time     `json:"generated_at"`
	Items       []ContentItem `json:"items"`
}

// ContentItem represents a single piece of generated news content with all metadata
type ContentItem struct {
	ID        string          `json:"id"`
	CreatedAt time.Time       `json:"created_at"`
	Source    SourceInfo      `json:"source"`
	Content   ContentInfo     `json:"content"`
	Media     MediaInfo       `json:"media"`
	SEO       SEOInfo         `json:"seo"`
	Platforms PlatformMetadata `json:"platforms"`
	Status    PostingStatus   `json:"posting_status"`
	Analytics AnalyticsInfo   `json:"analytics,omitempty"`
}

// SourceInfo contains information about the original news article
type SourceInfo struct {
	URL         string    `json:"url"`
	Title       string    `json:"title"`
	Author      string    `json:"author,omitempty"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	SourceName  string    `json:"source_name"`
}

// ContentInfo contains the generated content
type ContentInfo struct {
	Summary               string `json:"summary"`
	WordCount             int    `json:"word_count"`
	EstimatedDurationSecs int    `json:"estimated_duration_seconds"`
}

// MediaInfo contains paths and metadata for generated media files
type MediaInfo struct {
	AudioPath       string  `json:"audio_path"`
	VideoPath       string  `json:"video_path"`
	ThumbnailPath   *string `json:"thumbnail_path,omitempty"`
	DurationSeconds float64 `json:"duration_seconds"`
	Resolution      string  `json:"resolution"`
	AvatarID        string  `json:"avatar_id"`
}

// SEOInfo contains SEO and content categorization metadata
type SEOInfo struct {
	PrimaryKeywords   []string `json:"primary_keywords"`
	SecondaryKeywords []string `json:"secondary_keywords,omitempty"`
	Topics            []string `json:"topics"`
	Sentiment         string   `json:"sentiment,omitempty"`
	TargetAudience    []string `json:"target_audience,omitempty"`
}

// PlatformMetadata contains platform-specific metadata for all social platforms
type PlatformMetadata struct {
	YouTube   YouTubeMetadata   `json:"youtube"`
	TikTok    TikTokMetadata    `json:"tiktok"`
	Instagram InstagramMetadata `json:"instagram"`
	Twitter   TwitterMetadata   `json:"twitter"`
	Facebook  FacebookMetadata  `json:"facebook"`
	LinkedIn  LinkedInMetadata  `json:"linkedin"`
}

// YouTubeMetadata contains metadata optimized for YouTube
type YouTubeMetadata struct {
	Title           string           `json:"title"`            // Max 100 chars
	Description     string           `json:"description"`      // Max 5000 chars
	Tags            []string         `json:"tags"`             // Max 500 chars total
	CategoryID      string           `json:"category_id"`      // YouTube category (28 = Science & Technology)
	DefaultLanguage string           `json:"default_language"` // Default: "en"
	PrivacyStatus   string           `json:"privacy_status"`   // "public", "unlisted", "private"
	Timestamps      []VideoTimestamp `json:"timestamps,omitempty"`
}

// VideoTimestamp represents a timestamp marker in video description
type VideoTimestamp struct {
	Time  string `json:"time"`  // Format: "0:00"
	Label string `json:"label"` // Description of section
}

// TikTokMetadata contains metadata optimized for TikTok
type TikTokMetadata struct {
	Caption       string   `json:"caption"`        // Max 2200 chars (150 recommended)
	Hashtags      []string `json:"hashtags"`       // 3-5 recommended
	PrivacyLevel  string   `json:"privacy_level"`  // "public", "friends", "private"
	DuetEnabled   bool     `json:"duet_enabled"`   // Allow duets
	StitchEnabled bool     `json:"stitch_enabled"` // Allow stitches
}

// InstagramMetadata contains metadata optimized for Instagram Reels
type InstagramMetadata struct {
	Caption       string   `json:"caption"`   // Max 2200 chars
	Hashtags      []string `json:"hashtags"`  // 5-10 optimal (max 30)
	Location      *string  `json:"location"`  // Optional location tag
	Collaborators []string `json:"collaborators,omitempty"`
}

// TwitterMetadata contains metadata optimized for Twitter/X
type TwitterMetadata struct {
	Tweet         string   `json:"tweet"`          // Max 280 chars (more for media)
	Hashtags      []string `json:"hashtags"`       // 1-2 max recommended
	ReplySettings string   `json:"reply_settings"` // "everyone", "following", "mentioned"
}

// FacebookMetadata contains metadata optimized for Facebook
type FacebookMetadata struct {
	Message         string `json:"message"`          // Post text (max 63,206 chars)
	LinkDescription string `json:"link_description"` // Description for link preview
}

// LinkedInMetadata contains metadata optimized for LinkedIn
type LinkedInMetadata struct {
	PostText string   `json:"post_text"` // Max 3000 chars
	Hashtags []string `json:"hashtags"`  // Professional hashtags
}

// PostingStatus tracks which platforms content has been posted to
type PostingStatus struct {
	YouTube   PlatformStatus `json:"youtube"`
	TikTok    PlatformStatus `json:"tiktok"`
	Instagram PlatformStatus `json:"instagram"`
	Twitter   PlatformStatus `json:"twitter"`
	Facebook  PlatformStatus `json:"facebook"`
	LinkedIn  PlatformStatus `json:"linkedin"`
}

// PlatformStatus tracks posting status for a single platform
type PlatformStatus struct {
	Posted   bool       `json:"posted"`
	URL      *string    `json:"url,omitempty"`
	PostedAt *time.Time `json:"posted_at,omitempty"`
	Error    *string    `json:"error,omitempty"`
}

// AnalyticsInfo contains performance analytics (future use)
type AnalyticsInfo struct {
	Views       map[string]int `json:"views,omitempty"`
	Engagement  map[string]int `json:"engagement,omitempty"`
	LastUpdated *time.Time     `json:"last_updated,omitempty"`
}

// LLMMetadataResponse represents the structured response from LLM for metadata generation
type LLMMetadataResponse struct {
	Summary   string                     `json:"summary"`
	SEO       SEOInfo                    `json:"seo"`
	Platforms LLMPlatformMetadataResponse `json:"platforms"`
}

// LLMPlatformMetadataResponse is a simplified version for LLM output parsing
type LLMPlatformMetadataResponse struct {
	YouTube   LLMYouTubeMetadata   `json:"youtube"`
	TikTok    LLMTikTokMetadata    `json:"tiktok"`
	Instagram LLMInstagramMetadata `json:"instagram"`
	Twitter   LLMTwitterMetadata   `json:"twitter"`
	Facebook  LLMFacebookMetadata  `json:"facebook"`
	LinkedIn  LLMLinkedInMetadata  `json:"linkedin"`
}

// Simplified metadata structures for LLM output (easier parsing)
type LLMYouTubeMetadata struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Tags        []string         `json:"tags"`
	Timestamps  []VideoTimestamp `json:"timestamps,omitempty"`
}

type LLMTikTokMetadata struct {
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}

type LLMInstagramMetadata struct {
	Caption  string   `json:"caption"`
	Hashtags []string `json:"hashtags"`
}

type LLMTwitterMetadata struct {
	Tweet    string   `json:"tweet"`
	Hashtags []string `json:"hashtags"`
}

type LLMFacebookMetadata struct {
	Message         string `json:"message"`
	LinkDescription string `json:"link_description"`
}

type LLMLinkedInMetadata struct {
	PostText string   `json:"post_text"`
	Hashtags []string `json:"hashtags"`
}

