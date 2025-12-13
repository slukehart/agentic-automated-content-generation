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
	AudioPath  string `json:"audio_path"`            // Path to audio file (required)
	OutputPath string `json:"output_path"`           // Where to save the video
	AvatarID   string `json:"avatar_id,omitempty"`   // HeyGen avatar ID
	Background string `json:"background,omitempty"`  // Background color or image URL
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

// GenerateAvatarVideo creates an AI avatar video from an audio file
func GenerateAvatarVideo(audioPath string, outputPath string) (*VideoResponse, error) {
	return GenerateAvatarVideoWithOptions(audioPath, outputPath, "Kristin_public_3_20240108", "#0e1118")
}

// GenerateAvatarVideoWithOptions creates an avatar video with custom settings
func GenerateAvatarVideoWithOptions(audioPath, outputPath, avatarID, background string) (*VideoResponse, error) {
	request := VideoRequest{
		AudioPath:  audioPath,
		OutputPath: outputPath,
		AvatarID:   avatarID,
		Background: background,
	}

	return executeVideoGeneration(request)
}

// GenerateNewsVideo creates an AI avatar news video with audio narration
func GenerateNewsVideo(audioPath string, outputPath string) (*VideoResponse, error) {
	// Use professional news anchor avatar
	// Options:
	// - Kristin_public_3_20240108 (Female, professional)
	// - Wayne_20240711 (Male, professional)
	// See: https://app.heygen.com/avatars for more

	return GenerateAvatarVideoWithOptions(
		audioPath,
		outputPath,
		"Kristin_public_3_20240108", // Professional female news anchor
		"#0e1118",                     // Dark news studio background
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


