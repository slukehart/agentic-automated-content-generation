package video

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// VideoRequest represents a request to generate a video
type VideoRequest struct {
	Mode       string `json:"mode"`        // "text_to_video" or "image_to_video"
	Prompt     string `json:"prompt"`      // Text description
	ImagePath  string `json:"image_path,omitempty"`  // For image-to-video mode
	OutputPath string `json:"output_path"` // Where to save the video
	Model      string `json:"model,omitempty"`       // FAL model (default: fal-ai/ltx-video)
	Duration   int    `json:"duration,omitempty"`    // Duration in seconds (default: 5)
	FPS        int    `json:"fps,omitempty"`         // Frames per second (default: 24)
}

// VideoResponse represents the response from video generation
type VideoResponse struct {
	Status    string `json:"status"`
	VideoPath string `json:"video_path,omitempty"`
	VideoURL  string `json:"video_url,omitempty"`
	Duration  int    `json:"duration,omitempty"`
	FPS       int    `json:"fps,omitempty"`
	Message   string `json:"message,omitempty"`
	Details   string `json:"details,omitempty"`
}

// GenerateTextToVideo creates a video from a text prompt
func GenerateTextToVideo(prompt string, outputPath string) (*VideoResponse, error) {
	request := VideoRequest{
		Mode:       "text_to_video",
		Prompt:     prompt,
		OutputPath: outputPath,
		Duration:   5,
		FPS:        24,
	}

	return executeVideoGeneration(request)
}

// GenerateImageToVideo creates a video from an image with animation
func GenerateImageToVideo(imagePath string, prompt string, outputPath string) (*VideoResponse, error) {
	request := VideoRequest{
		Mode:       "image_to_video",
		ImagePath:  imagePath,
		Prompt:     prompt,
		OutputPath: outputPath,
		Duration:   5,
	}

	return executeVideoGeneration(request)
}

// GenerateNewsVideo creates a video for a news summary
func GenerateNewsVideo(newsSummary string, outputPath string) (*VideoResponse, error) {
	// Create a cinematic prompt from the news summary
	prompt := fmt.Sprintf(
		"Generate a video from the following news summary that is read by a female news anchor with brunette hair and dark tan skin: %s",
		truncateString(newsSummary, 200), // FAL has prompt length limits
	)

	return GenerateTextToVideo(prompt, outputPath)
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

	// Pass environment variables to subprocess (critical for FAL_KEY)
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))
	cmd.Env = append(cmd.Env, "FAL_KEY="+os.Getenv("FAL_KEY"))
	// Add any other API keys you might need
	if newsKey := os.Getenv("NEWS_API_KEY"); newsKey != "" {
		cmd.Env = append(cmd.Env, "NEWS_API_KEY="+newsKey)
	}
	if grokKey := os.Getenv("GROK_API_KEY"); grokKey != "" {
		cmd.Env = append(cmd.Env, "GROK_API_KEY="+grokKey)
	}

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

// truncateString truncates a string to maxLen characters
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

