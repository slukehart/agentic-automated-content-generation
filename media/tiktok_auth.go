package media

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"time"

	"golang.org/x/oauth2"
)

const (
	// TikTok OAuth endpoints
	tiktokAuthURL  = "https://www.tiktok.com/v2/auth/authorize/"
	tiktokTokenURL = "https://open.tiktokapis.com/v2/oauth/token/"

	// Credential cache
	tiktokCredentialsCacheDir = ".credentials"
	tiktokTokenCacheFilename  = "tiktok-oauth.json"
	tiktokCallbackPort        = "8080"

	// Redirect URI - must match EXACTLY what is registered in TikTok Developer Portal
	tiktokRedirectURI = "https://slukehart.github.io/agentic-automated-content-generation/callback"
)

// TikTokConfig holds TikTok API configuration
type TikTokConfig struct {
	ClientKey    string
	ClientSecret string
	Sandbox      bool
}

// GetTikTokConfig reads TikTok credentials from environment
// If sandbox is true, uses TIKTOK_CLIENT_KEY_SANDBOX and TIKTOK_CLIENT_SECRET_SANDBOX
// Otherwise uses TIKTOK_CLIENT_KEY and TIKTOK_CLIENT_SECRET
func GetTikTokConfig(sandbox bool) (*TikTokConfig, error) {
	var clientKey, clientSecret string

	if sandbox {
		clientKey = os.Getenv("TIKTOK_CLIENT_KEY_SANDBOX")
		clientSecret = os.Getenv("TIKTOK_CLIENT_SECRET_SANDBOX")
		if clientKey == "" || clientSecret == "" {
			return nil, fmt.Errorf("TIKTOK_CLIENT_KEY_SANDBOX and TIKTOK_CLIENT_SECRET_SANDBOX must be set in .env for sandbox mode")
		}
	} else {
		clientKey = os.Getenv("TIKTOK_CLIENT_KEY")
		clientSecret = os.Getenv("TIKTOK_CLIENT_SECRET")
		if clientKey == "" || clientSecret == "" {
			return nil, fmt.Errorf("TIKTOK_CLIENT_KEY and TIKTOK_CLIENT_SECRET must be set in .env")
		}
	}

	return &TikTokConfig{
		ClientKey:    clientKey,
		ClientSecret: clientSecret,
		Sandbox:      sandbox,
	}, nil
}

// GetTikTokAccessToken returns a valid TikTok access token
// On first run, it will open a browser for OAuth authorization
// Subsequent runs will use cached credentials from ~/.credentials/tiktok-oauth.json
func GetTikTokAccessToken(ctx context.Context, sandbox bool) (string, error) {
	config, err := GetTikTokConfig(sandbox)
	if err != nil {
		return "", err
	}

	// Try to load cached token
	cacheFile, err := tiktokTokenCacheFile()
	if err != nil {
		return "", fmt.Errorf("unable to get path to cached credential file: %w", err)
	}

	token, err := tiktokTokenFromFile(cacheFile)
	if err == nil && token.Valid() {
		// Token is valid, return it
		return token.AccessToken, nil
	}

	// Need to get new token via OAuth flow
	fmt.Println("\n🔐 TikTok OAuth Authorization Required")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	if sandbox {
		fmt.Println("Running in SANDBOX mode (for demo/testing)")
	}

	// Debug: Show credential status (masked)
	if len(config.ClientKey) > 0 {
		fmt.Printf("✓ Client Key loaded: %s...%s\n", config.ClientKey[:4], config.ClientKey[len(config.ClientKey)-4:])
	} else {
		fmt.Println("✗ Client Key is EMPTY!")
		return "", fmt.Errorf("client_key is empty - check your .env file")
	}
	fmt.Println()

	// Build OAuth config
	// Note: TikTok uses "client_key" instead of standard "client_id"
	// Note: TikTok requires comma-separated scopes, NOT space-separated (OAuth2 default).
	// We pass scopes manually via SetAuthURLParam in getTikTokTokenFromWeb instead.
	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientKey,
		ClientSecret: config.ClientSecret,
		RedirectURL:  tiktokRedirectURI,
		Scopes:       []string{}, // Scopes set explicitly in auth URL (comma-separated for TikTok)
		Endpoint: oauth2.Endpoint{
			AuthURL:  tiktokAuthURL,
			TokenURL: tiktokTokenURL,
		},
	}

	// Get token from web
	token, err = getTikTokTokenFromWeb(oauthConfig)
	if err != nil {
		return "", fmt.Errorf("failed to get token from web: %w", err)
	}

	// Save token for future use
	saveTikTokToken(cacheFile, token)

	return token.AccessToken, nil
}

// getTikTokTokenFromWeb uses OAuth to request a Token via browser with PKCE
func getTikTokTokenFromWeb(config *oauth2.Config) (*oauth2.Token, error) {
	// Generate PKCE parameters (required by TikTok)
	codeVerifier, err := generateCodeVerifier()
	if err != nil {
		return nil, fmt.Errorf("failed to generate code verifier: %w", err)
	}

	codeChallenge := generateCodeChallenge(codeVerifier)

	// Generate authorization URL with PKCE parameters
	// Note: TikTok uses "client_key" instead of standard "client_id"
	// Note: TikTok requires comma-separated scopes, not space-separated
	authURL := config.AuthCodeURL(
		"state-token",
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("client_key", config.ClientID),
		oauth2.SetAuthURLParam("scope", "user.info.basic,video.upload,video.publish"),
		oauth2.SetAuthURLParam("code_challenge", codeChallenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Printf("1. Open this URL in your browser:\n\n")
	fmt.Printf("   %v\n\n", authURL)
	fmt.Printf("2. Authorize the application\n")
	fmt.Printf("3. You'll be redirected to localhost (may show an error page)\n")
	fmt.Printf("4. Copy the 'code' parameter from the URL\n")
	fmt.Printf("5. Paste it below and press Enter\n\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("\nEnter authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("unable to read authorization code: %w", err)
	}

	// Exchange code for token with PKCE code_verifier
	// TikTok also needs client_key in token exchange
	token, err := config.Exchange(
		context.Background(),
		code,
		oauth2.SetAuthURLParam("code_verifier", codeVerifier),
		oauth2.SetAuthURLParam("client_key", config.ClientID),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve token from web: %w", err)
	}

	return token, nil
}

// generateCodeVerifier creates a random string for PKCE
func generateCodeVerifier() (string, error) {
	// Generate 32 random bytes (256 bits)
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// Base64 URL encode (without padding)
	return base64.RawURLEncoding.EncodeToString(bytes), nil
}

// generateCodeChallenge creates SHA256 hash of code_verifier for PKCE
func generateCodeChallenge(verifier string) string {
	hash := sha256.Sum256([]byte(verifier))
	// Base64 URL encode (without padding)
	return base64.RawURLEncoding.EncodeToString(hash[:])
}

// tiktokTokenCacheFile generates credential file path/filename
// Returns: ~/.credentials/tiktok-oauth.json
func tiktokTokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	tokenCacheDir := filepath.Join(usr.HomeDir, tiktokCredentialsCacheDir)

	// Create directory if it doesn't exist
	if err := os.MkdirAll(tokenCacheDir, 0700); err != nil {
		return "", fmt.Errorf("failed to create credentials directory: %w", err)
	}

	return filepath.Join(tokenCacheDir, tiktokTokenCacheFilename), nil
}

// tiktokTokenFromFile retrieves a Token from a given file path
func tiktokTokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	return t, err
}

// saveTikTokToken uses a file path to create a file and store the token in it
func saveTikTokToken(file string, token *oauth2.Token) {
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

// RefreshTikTokToken refreshes an expired token
func RefreshTikTokToken(ctx context.Context, refreshToken string, sandbox bool) (*oauth2.Token, error) {
	config, err := GetTikTokConfig(sandbox)
	if err != nil {
		return nil, err
	}

	oauthConfig := &oauth2.Config{
		ClientID:     config.ClientKey,
		ClientSecret: config.ClientSecret,
		Endpoint: oauth2.Endpoint{
			TokenURL: tiktokTokenURL,
		},
	}

	token := &oauth2.Token{
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(-time.Hour), // Force refresh
	}

	tokenSource := oauthConfig.TokenSource(ctx, token)
	newToken, err := tokenSource.Token()
	if err != nil {
		return nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Save refreshed token
	cacheFile, err := tiktokTokenCacheFile()
	if err != nil {
		return nil, fmt.Errorf("failed to resolve token cache path: %w", err)
	}
	saveTikTokToken(cacheFile, newToken)

	return newToken, nil
}

