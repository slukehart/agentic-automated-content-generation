// Package video generates the avatar news video by submitting a job to the
// LongCat-Video-Avatar-1.5 RunPod Serverless endpoint (see video/longcat-runpod).
package video

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const (
	// DefaultAvatarID identifies the avatar pipeline recorded in the manifest.
	DefaultAvatarID = "longcat-video-avatar-1.5"

	// DefaultPrompt is the scene description passed to LongCat alongside the image.
	DefaultPrompt = "A professional news anchor sits at a newsroom desk, speaking directly to camera " +
		"with natural expressions and subtle head movement."

	runpodBaseURL = "https://api.runpod.ai/v2"

	// LongCat renders 93 frames @ 25fps for the first segment. Each later
	// segment reuses 13 frames as conditioning, so it adds only 80 new frames.
	firstSegmentSeconds = 93.0 / 25
	nextSegmentSeconds  = 80.0 / 25

	pollInterval = 15 * time.Second
	// A100 queue waits (12-26min) plus ~42min of compute for a 60s video.
	jobTimeout = 2 * time.Hour
)

// VideoResponse describes the outcome of a video generation job.
type VideoResponse struct {
	Status   string // "success" or "error"
	Message  string
	Duration float64 // seconds
	JobID    string
	Applied  map[string]any // inference args the handler reports it used
}

// Options are the optional per-job inference settings. A nil field leaves the
// handler's (and the script's) default in place.
type Options struct {
	Prompt         string
	RefImgIndex    *int // lower anchors harder to the reference; script default 10
	MaskFrameRange *int // larger reduces repeated motion; script default 3
}

// GenerateNarration synthesizes text to a WAV file with Kokoro by shelling out
// to cmd/testttsprompt/generate_audio.py. This is the local dev path; it is to
// be replaced by a serverless TTS step.
func GenerateNarration(text, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("create audio directory: %w", err)
	}
	cmd := exec.Command(
		filepath.Join(".venv", "bin", "python3"),
		filepath.Join("cmd", "testttsprompt", "generate_audio.py"),
		text, outputPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("kokoro narration failed: %w", err)
	}
	return nil
}

// GenerateAvatarVideo submits imagePath (avatar with background already
// composited in) and audioPath (narration WAV) to the RunPod endpoint, waits
// for the job, and writes the resulting MP4 to outputPath.
func GenerateAvatarVideo(imagePath, audioPath, prompt, outputPath string) (*VideoResponse, error) {
	return GenerateAvatarVideoWithOptions(imagePath, audioPath, outputPath, Options{Prompt: prompt})
}

// GenerateAvatarVideoWithOptions is GenerateAvatarVideo with per-job settings.
func GenerateAvatarVideoWithOptions(imagePath, audioPath, outputPath string, opts Options) (*VideoResponse, error) {
	prompt := opts.Prompt
	apiKey := os.Getenv("RUNPOD_API_KEY")
	endpointID := os.Getenv("RUNPOD_ENDPOINT_ID")
	if apiKey == "" || endpointID == "" {
		return nil, fmt.Errorf("RUNPOD_API_KEY and RUNPOD_ENDPOINT_ID must be set")
	}
	if prompt == "" {
		prompt = DefaultPrompt
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}
	audioBytes, err := os.ReadFile(audioPath)
	if err != nil {
		return nil, fmt.Errorf("read audio: %w", err)
	}
	duration, err := wavDuration(audioBytes)
	if err != nil {
		return nil, fmt.Errorf("read narration duration: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
	defer cancel()

	input := map[string]any{
		"cond_image":   base64.StdEncoding.EncodeToString(imageBytes),
		"cond_audio":   base64.StdEncoding.EncodeToString(audioBytes),
		"prompt":       prompt,
		"num_segments": segmentsFor(duration),
	}
	if opts.RefImgIndex != nil {
		input["ref_img_index"] = *opts.RefImgIndex
	}
	if opts.MaskFrameRange != nil {
		input["mask_frame_range"] = *opts.MaskFrameRange
	}
	payload := map[string]any{"input": input}
	var submitted struct {
		ID string `json:"id"`
	}
	if err := runpodRequest(ctx, http.MethodPost, fmt.Sprintf("%s/%s/run", runpodBaseURL, endpointID), apiKey, payload, &submitted); err != nil {
		return nil, fmt.Errorf("submit job: %w", err)
	}
	fmt.Printf("    📤 RunPod job submitted: %s\n", submitted.ID)
	resp := &VideoResponse{Status: "error", Duration: duration, JobID: submitted.ID}

	statusURL := fmt.Sprintf("%s/%s/status/%s", runpodBaseURL, endpointID, submitted.ID)
	for {
		var job struct {
			Status string `json:"status"`
			Error  string `json:"error"`
			Output struct {
				VideoB64 string         `json:"video_b64"`
				Error    string         `json:"error"`
				Applied  map[string]any `json:"applied"`
			} `json:"output"`
		}
		if err := runpodRequest(ctx, http.MethodGet, statusURL, apiKey, nil, &job); err != nil {
			return resp, fmt.Errorf("poll job %s: %w", submitted.ID, err)
		}

		switch job.Status {
		case "COMPLETED":
			if job.Output.Error != "" {
				resp.Message = job.Output.Error
				return resp, fmt.Errorf("job %s failed", submitted.ID)
			}
			videoBytes, err := base64.StdEncoding.DecodeString(job.Output.VideoB64)
			if err != nil || len(videoBytes) == 0 {
				return resp, fmt.Errorf("job %s returned no decodable video", submitted.ID)
			}
			if err := os.WriteFile(outputPath, videoBytes, 0644); err != nil {
				return resp, fmt.Errorf("write video: %w", err)
			}
			resp.Applied = job.Output.Applied
			resp.Status = "success"
			return resp, nil
		case "FAILED", "CANCELLED", "TIMED_OUT":
			resp.Message = job.Error
			return resp, fmt.Errorf("job %s ended with status %s", submitted.ID, job.Status)
		}

		select {
		case <-ctx.Done():
			return resp, fmt.Errorf("job %s did not finish within %s (last status %s)", submitted.ID, jobTimeout, job.Status)
		case <-time.After(pollInterval):
		}
	}
}

// segmentsFor returns how many LongCat segments cover seconds of narration.
func segmentsFor(seconds float64) int {
	if seconds <= firstSegmentSeconds {
		return 1
	}
	return 1 + int(math.Ceil((seconds-firstSegmentSeconds)/nextSegmentSeconds))
}

func runpodRequest(ctx context.Context, method, url, apiKey string, body any, out any) error {
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", res.StatusCode, truncate(string(data), 300))
	}
	return json.Unmarshal(data, out)
}

// wavDuration returns the playback length in seconds of a PCM WAV file.
func wavDuration(data []byte) (float64, error) {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return 0, fmt.Errorf("not a WAV file")
	}
	var byteRate uint32
	for pos := 12; pos+8 <= len(data); {
		id := string(data[pos : pos+4])
		size := int(binary.LittleEndian.Uint32(data[pos+4 : pos+8]))
		body := pos + 8
		switch id {
		case "fmt ":
			if body+12 > len(data) {
				return 0, fmt.Errorf("truncated fmt chunk")
			}
			byteRate = binary.LittleEndian.Uint32(data[body+8 : body+12])
		case "data":
			if byteRate == 0 {
				return 0, fmt.Errorf("data chunk before fmt chunk")
			}
			if body+size > len(data) {
				size = len(data) - body // streamed/truncated header
			}
			return float64(size) / float64(byteRate), nil
		}
		pos = body + size + size%2
	}
	return 0, fmt.Errorf("no data chunk")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
