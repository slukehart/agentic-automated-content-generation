package audio

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// TTSRequest represents a request to generate audio from text using Google Cloud TTS
type TTSRequest struct {
	Text       string  `json:"text"`                 // Text to convert to speech
	OutputPath string  `json:"output_path"`          // Where to save the audio file
	Voice      string  `json:"voice,omitempty"`      // Voice ID (Google Cloud voice name)
	Speed      float64 `json:"speed,omitempty"`      // Speech speed (0.5 to 2.0)
}

// TTSResponse represents the response from audio generation
type TTSResponse struct {
	Status    string  `json:"status"`
	AudioPath string  `json:"audio_path,omitempty"`
	AudioURL  string  `json:"audio_url,omitempty"`
	Duration  float64 `json:"duration,omitempty"` // Audio duration in seconds
	Message   string  `json:"message,omitempty"`
	Details   string  `json:"details,omitempty"`
}

// GenerateAudio creates audio from text using Google Cloud TTS
func GenerateAudio(text string, outputPath string) (*TTSResponse, error) {
	return GenerateAudioWithOptions(text, outputPath, "en-US-Neural2-J", 1.0)
}

// GenerateAudioWithOptions creates audio with custom voice and speed settings
func GenerateAudioWithOptions(text, outputPath, voice string, speed float64) (*TTSResponse, error) {
	request := TTSRequest{
		Text:       text,
		OutputPath: outputPath,
		Voice:      voice,
		Speed:      speed,
	}

	return executeTTSGeneration(request)
}

// GenerateNewsAudio creates professional news narration audio
func GenerateNewsAudio(newsSummary string, outputPath string) (*TTSResponse, error) {
	// Use a professional, clear voice for news (Google Cloud TTS)
	// en-US-Neural2-J = Male professional voice
	// en-US-Neural2-F = Female professional voice
	return GenerateAudioWithOptions(
		newsSummary,
		outputPath,
		"en-US-Neural2-F", // Professional male voice
		1.0,               // Normal speed
	)
}

// executeTTSGeneration handles the actual Python script execution
func executeTTSGeneration(request TTSRequest) (*TTSResponse, error) {
	// Marshal request to JSON
	jsonData, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Get the path to the Python script
	scriptPath := filepath.Join("audio", "tts_generation.py")

	// Execute Python script with Poetry
	cmd := exec.Command("poetry", "run", "python", scriptPath)
	cmd.Stdin = bytes.NewReader(jsonData)

	// Pass environment variables to subprocess
	cmd.Env = append(cmd.Env, "PATH="+os.Getenv("PATH"))

	// Google Cloud credentials
	if gcpCreds := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); gcpCreds != "" {
		cmd.Env = append(cmd.Env, "GOOGLE_APPLICATION_CREDENTIALS="+gcpCreds)
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
	var response TTSResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w\nOutput: %s", err, stdout.String())
	}

	// Check for errors in response
	if response.Status != "success" {
		return &response, fmt.Errorf("audio generation failed: %s", response.Message)
	}

	return &response, nil
}

// MergeAudioVideo combines an audio file and video file using ffmpeg
func MergeAudioVideo(videoPath, audioPath, outputPath string) error {
	// Check if ffmpeg is available
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found. Please install ffmpeg: brew install ffmpeg")
	}

	// Merge audio and video using ffmpeg
	// -y: overwrite output file
	// -i: input files
	// -c:v copy: copy video codec (no re-encoding)
	// -c:a aac: encode audio as AAC
	// -shortest: finish when shortest stream ends
	cmd := exec.Command("ffmpeg",
		"-y",
		"-i", videoPath,
		"-i", audioPath,
		"-c:v", "copy",
		"-c:a", "aac",
		"-strict", "experimental",
		"-shortest",
		outputPath,
	)

	// Capture output for debugging
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %w\nStderr: %s", err, stderr.String())
	}

	return nil
}

