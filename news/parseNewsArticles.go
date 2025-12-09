package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	grok "github.com/SimonMorphy/grok-go"
	"github.com/joho/godotenv"
)

type Article struct {
	Source Source `json:"source"`
	Title string `json:"title"`
	Author string `json:"author"`
	Description string `json:"description"`
	Url string `json:"url"`
	UrlToImage string `json:"urlToImage"`
	PublishedAt string `json:"publishedAt"`
	Content string `json:"content"`
}

type Source struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

type NewsAPIResponse struct {
	Status string `json:"status"`
	TotalResults int `json:"totalResults"`
	Articles []Article `json:"articles"`
}


type AiArticleParameters struct {
	ArticleUrl string `json:"articleUrl"`
	ArticleTitle string `json:"articleTitle"`
}

// ParseNewsArticles fetches top headlines from NewsAPI
func ParseNewsArticles() AiArticleParameters {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("NEWS_API_KEY")

	url := "https://newsapi.org/v2/top-headlines?country=us&sortBy=popularity&apiKey=" + apiKey

	resp, err := http.Get(url)
	if err != nil {
		log.Fatal("Error getting news articles:", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal("Error reading news articles:", err)
	}

	var apiResponse NewsAPIResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		log.Fatal("Error unmarshaling JSON:", err)
	}

	var newsReports []AiArticleParameters

	for _, article := range apiResponse.Articles {
		newsReports = append(newsReports, AiArticleParameters{
			ArticleUrl: article.Url,
			ArticleTitle: article.Title,
		})
	}

	return newsReports[0]
}


// GenerateBatchNewsReportSummaries - MOST TOKEN EFFICIENT
// Processes ALL articles in a single API call with prompt sent only once
func GenerateBatchNewsReportSummaries(articles AiArticleParameters, systemPrompt string) ([]string, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	apiKey := os.Getenv("X_AI_KEY")
	if apiKey == "" {
		log.Fatal("X_AI_KEY environment variable not set")
	}

	// Initialize client with extended timeout for batch processing
	// Detailed summaries (150-200 words each) take longer to generate
	client, err := grok.NewClientWithOptions(apiKey,
		grok.WithTimeout(5*time.Minute), // 5 minute timeout for batch processing
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}

	// Build a single message with all articles
	var articlesText string
	articlesText += fmt.Sprintf("\n\nArticle %d:\nTitle: %s\nURL: %s", 1, articles.ArticleTitle, articles.ArticleUrl)

	fmt.Printf("Processing 1 article with Grok (this may take 1-2 minutes for detailed summaries)...\n")

	request := &grok.ChatCompletionRequest{
		Model: "grok-3",
		Messages: []grok.ChatCompletionMessage{
			{
				Role:    "system",
				Content: systemPrompt + "\n\nPlease process each article and return the summaries in a numbered list format (1., 2., 3., etc.).",
			},
			{
				Role:    "user",
				Content: articlesText,
			},
		},
		Temperature: 0.7,
		MaxTokens:   8000, // ~200 words per article × 10 articles × 1.33 (word-to-token ratio) ≈ 2,660 tokens + buffer
	}
	// Set stream_options to satisfy API requirements
	request.StreamOptions.IncludeUsage = true

	// Send request with extended context timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	response, err := grok.CreateChatCompletion(ctx, client, request)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat completion: %v", err)
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response from Grok")
	}

	// For now, return the entire response as a single string
	// You could parse this to extract individual summaries if needed
	return []string{response.Choices[0].Message.Content}, nil
}
