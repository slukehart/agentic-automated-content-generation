package main

import (
	"fmt"
	"log"
	"os"

	"content-generation-automation/news"
)

func main() {
	prompt := os.Getenv("TEST_PROMPT")

	result, err := news.GenerateNewsroomBackground(prompt)
	if err != nil {
		log.Fatalf("generation failed: %v", err)
	}

	fmt.Printf("saved to: %s\n", result.ImagePath)
}
