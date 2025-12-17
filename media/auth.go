package media

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
)

const (
	// clientSecretFile is the OAuth2 client credentials file
	clientSecretFile = "client_secret.json"
	// credentialsCacheDir is where OAuth tokens are cached
	credentialsCacheDir = ".credentials"
	// tokenCacheFilename is the filename for cached YouTube OAuth token
	tokenCacheFilename = "youtube-oauth.json"
)

// GetYouTubeClient returns an authenticated YouTube service client
// On first run, it will open a browser for OAuth authorization
// Subsequent runs will use cached credentials from ~/.credentials/youtube-oauth.json
func GetYouTubeClient(ctx context.Context) (*youtube.Service, error) {
	// Read client secret file
	b, err := os.ReadFile(clientSecretFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read client secret file: %w\nMake sure %s exists in project root", err, clientSecretFile)
	}

	// Parse OAuth config from client secret
	// Request youtube.upload scope for uploading videos
	config, err := google.ConfigFromJSON(b, youtube.YoutubeUploadScope)
	if err != nil {
		return nil, fmt.Errorf("unable to parse client secret file to config: %w", err)
	}

	// Get authenticated HTTP client
	client := getClient(ctx, config)

	// Create YouTube service
	service, err := youtube.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("error creating YouTube client: %w", err)
	}

	return service, nil
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file: %v", err)
	}

	// Try to load token from cache
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		// No cached token, need to authorize via browser
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}

	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token via browser OAuth flow
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)

	fmt.Printf("\n🔐 YouTube OAuth Authorization Required\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
	fmt.Printf("This is a one-time setup. Your credentials will be cached for future use.\n\n")
	fmt.Printf("1. Open this URL in your browser:\n\n")
	fmt.Printf("   %v\n\n", authURL)
	fmt.Printf("2. Authorize the application\n")
	fmt.Printf("3. Copy the authorization code\n")
	fmt.Printf("4. Paste it below and press Enter\n\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\nEnter authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}

	return tok
}

// tokenCacheFile generates credential file path/filename
// Returns: ~/.credentials/youtube-oauth.json
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, credentialsCacheDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(tokenCacheDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create credentials directory: %w", err)
	}

	return filepath.Join(tokenCacheDir, url.QueryEscape(tokenCacheFilename)), nil
}

// tokenFromFile retrieves a Token from a given file path
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// saveToken uses a file path to create a file and store the token in it
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("\n✅ Saving credentials to: %s\n", file)
	fmt.Printf("   (Future runs will use cached credentials automatically)\n\n")

	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(token); err != nil {
		log.Fatalf("Unable to encode oauth token: %v", err)
	}
}

