package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// This is a local dev/test harness only. Production runs the same Kokoro
// inference logic (see generate_audio.py) inside a RunPod Serverless CPU
// handler, not as a Go subprocess.
func main() {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	text := os.Getenv("TEST_TEXT")
	if text == "" {
		text = "Breaking news tonight: markets rallied after the central bank held interest rates steady, " +
			"signaling confidence in the economy's continued recovery."
	}

	if err := os.MkdirAll("audio", 0755); err != nil {
		log.Fatalf("failed to create audio directory: %v", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	outputPath := filepath.Join("audio", fmt.Sprintf("narration_%s.wav", timestamp))

	scriptPath := filepath.Join("cmd", "testttsprompt", "generate_audio.py")
	venvPython := filepath.Join(".venv", "bin", "python3")

	cmd := exec.Command(venvPython, scriptPath, text, outputPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatalf("generation failed: %v", err)
	}

	fmt.Printf("saved to: %s\n", outputPath)
}
