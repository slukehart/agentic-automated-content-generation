package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoRequest represents a request to generate an avatar video with HeyGen
type VideoRequest struct {
	Text            string `json:"text,omitempty"`             // Text for TTS (preferred method)
	AudioPath       string `json:"audio_path,omitempty"`       // Path to audio file (legacy)
	OutputPath      string `json:"output_path"`                // Where to save the video
	AvatarID        string `json:"avatar_id,omitempty"`        // HeyGen avatar ID
	VoiceID         string `json:"voice_id,omitempty"`         // HeyGen voice ID for TTS
	Background      string `json:"background,omitempty"`       // Background type ("newsroom", "#hex", or "image")
	BackgroundImage string `json:"background_image,omitempty"` // URL or asset ID for custom background
}

// VideoResponse represents the response from video generation
type VideoResponse struct {
	Status    string  `json:"status"`
	VideoPath string  `json:"video_path,omitempty"`
	VideoURL  string  `json:"video_url,omitempty"`
	VideoID   string  `json:"video_id,omitempty"`   // HeyGen video ID
	Duration  float64 `json:"duration,omitempty"`   // Video duration in seconds
	Message   string  `json:"message,omitempty"`
	Details   string  `json:"details,omitempty"`
}

// GenerateNewsVideoFromText creates an AI avatar news video using text-to-speech
// This is the RECOMMENDED method - no separate audio generation needed!
func GenerateNewsVideoFromText(text string, outputPath string) (*VideoResponse, error) {
	return GenerateNewsVideoFromTextWithOptions(
		text,
		outputPath,
		DefaultAvatarID,
		DefaultVoiceID,
		DefaultBackground,
	)
}

// GenerateNewsVideoFromTextWithOptions creates a news video with custom avatar, voice, and background
func GenerateNewsVideoFromTextWithOptions(text, outputPath, avatarID, voiceID, background string) (*VideoResponse, error) {
	request := VideoRequest{
		Text:       text,
		OutputPath: outputPath,
		AvatarID:   avatarID,
		VoiceID:    voiceID,
		Background: background,
	}

	return executeVideoGeneration(request)
}

// GenerateNewsVideoWithBackgroundImage creates a news video with a custom background image
func GenerateNewsVideoWithBackgroundImage(text, outputPath, backgroundImagePath string) (*VideoResponse, error) {
	request := VideoRequest{
		Text:            text,
		OutputPath:      outputPath,
		AvatarID:        DefaultAvatarID,
		VoiceID:         DefaultVoiceID,
		Background:      "image",
		BackgroundImage: backgroundImagePath,
	}

	return executeVideoGeneration(request)
}

// GenerateNewsVideoWithAllOptions creates a news video with full customization including background image
func GenerateNewsVideoWithAllOptions(text, outputPath, avatarID, voiceID, backgroundImagePath string, generateCaptions, burnInCaptions bool) (*VideoResponse, error) {
	request := VideoRequest{
		Text:            text,
		OutputPath:      outputPath,
		AvatarID:        avatarID,
		VoiceID:         voiceID,
		Background:      "image",
		BackgroundImage: backgroundImagePath,
	}

	return executeVideoGeneration(request)
}

// GenerateAvatarVideo creates an AI avatar video from an audio file (LEGACY)
// Use GenerateNewsVideoFromText instead for simpler pipeline
func GenerateAvatarVideo(audioPath string, outputPath string) (*VideoResponse, error) {
	return GenerateAvatarVideoWithOptions(audioPath, outputPath, DefaultAvatarID, DefaultBackgroundColor)
}

// GenerateAvatarVideoWithOptions creates an avatar video with custom settings (LEGACY)
func GenerateAvatarVideoWithOptions(audioPath, outputPath, avatarID, background string) (*VideoResponse, error) {
	request := VideoRequest{
		AudioPath:  audioPath,
		OutputPath: outputPath,
		AvatarID:   avatarID,
		Background: background,
	}

	return executeVideoGeneration(request)
}

// GenerateNewsVideo creates an AI avatar news video with audio narration (LEGACY)
// Use GenerateNewsVideoFromText instead
func GenerateNewsVideo(audioPath string, outputPath string) (*VideoResponse, error) {
	// Use professional news anchor avatar
	// See video/constants.go to change default avatar
	// Or visit: https://app.heygen.com/avatars for more options

	return GenerateAvatarVideoWithOptions(
		audioPath,
		outputPath,
		DefaultAvatarID,
		DefaultBackground,
	)
}

// executeVideoGeneration handles the actual Python script execution
func executeVideoGeneration(request VideoRequest) (*VideoResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Get the path to the Python script
	scriptPath := filepath.Join("video", "video_generation.py")

	// Execute Python script with Poetry
	cmd := exec.Command("poetry", "run", "python", scriptPath)
	cmd.Stdin = bytes.NewReader(jsonData)

	// Pass environment variables to subprocess (critical for HEYGEN_API_KEY)
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "HEYGEN_API_KEY="+os.Getenv("HEYGEN_API_KEY"))

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("python script failed: %w\nStderr: %s", err, stderr.String())
	}

	// Parse response
	var response VideoResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nOutput: %s", err, stdout.String())
	}

	// Check for errors in response
	if response.Status != "success" {
		return &response, fmt.Errorf("video generation failed: %s", response.Message)
	}

	return &response, nil
}


