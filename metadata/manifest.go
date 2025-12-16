package metadata

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	// ManifestFileName is the default name for the content manifest file
	ManifestFileName = "content_manifest.json"
	// ManifestVersion is the current version of the manifest format
	ManifestVersion = "1.0"
)

// ManifestManager handles CRUD operations for the content manifest
type ManifestManager struct {
	filePath string
	mu       sync.RWMutex // Protect concurrent access
}

// NewManifestManager creates a new manifest manager
func NewManifestManager(filePath string) *ManifestManager {
	if filePath == "" {
		filePath = ManifestFileName
	}
	return &ManifestManager{
		filePath: filePath,
	}
}

// CreateManifest creates a new empty manifest file
func (m *ManifestManager) CreateManifest() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	manifest := ContentManifest{
		Version:     ManifestVersion,
		GeneratedAt: time.Now(),
		Items:       []ContentItem{},
	}

	return m.saveManifestUnsafe(manifest)
}

// LoadManifest loads the manifest from disk, creates new if doesn't exist
func (m *ManifestManager) LoadManifest() (*ContentManifest, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if file exists
	if _, err := os.Stat(m.filePath); os.IsNotExist(err) {
		// Return empty manifest, will be created on first save
		return &ContentManifest{
			Version:     ManifestVersion,
			GeneratedAt: time.Now(),
			Items:       []ContentItem{},
		}, nil
	}

	// Read file
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest file: %w", err)
	}

	// Parse JSON
	var manifest ContentManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, fmt.Errorf("failed to parse manifest JSON: %w", err)
	}

	return &manifest, nil
}

// SaveManifest saves the manifest to disk
func (m *ManifestManager) SaveManifest(manifest ContentManifest) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.saveManifestUnsafe(manifest)
}

// saveManifestUnsafe saves without locking (internal use when already locked)
func (m *ManifestManager) saveManifestUnsafe(manifest ContentManifest) error {
	// Update generated_at timestamp
	manifest.GeneratedAt = time.Now()

	// Marshal to JSON with indentation for readability
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal manifest: %w", err)
	}

	// Write to file
	if err := os.WriteFile(m.filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write manifest file: %w", err)
	}

	return nil
}

// AddItem adds a new content item to the manifest
func (m *ManifestManager) AddItem(item ContentItem) error {
	manifest, err := m.LoadManifest()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Check for duplicate ID
	for _, existing := range manifest.Items {
		if existing.ID == item.ID {
			return fmt.Errorf("item with ID %s already exists", item.ID)
		}
	}

	// Add item
	manifest.Items = append(manifest.Items, item)

	// Save
	return m.SaveManifest(*manifest)
}

// UpdateItem updates an existing item in the manifest
func (m *ManifestManager) UpdateItem(item ContentItem) error {
	manifest, err := m.LoadManifest()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Find and update item
	found := false
	for i, existing := range manifest.Items {
		if existing.ID == item.ID {
			manifest.Items[i] = item
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item with ID %s not found", item.ID)
	}

	// Save
	return m.SaveManifest(*manifest)
}

// GetItem retrieves a single item by ID
func (m *ManifestManager) GetItem(id string) (*ContentItem, error) {
	manifest, err := m.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	for _, item := range manifest.Items {
		if item.ID == id {
			return &item, nil
		}
	}

	return nil, fmt.Errorf("item with ID %s not found", id)
}

// GetAllItems retrieves all items from the manifest
func (m *ManifestManager) GetAllItems() ([]ContentItem, error) {
	manifest, err := m.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	return manifest.Items, nil
}

// GetItemsByStatus retrieves items filtered by posting status on a platform
func (m *ManifestManager) GetItemsByStatus(platform string, posted bool) ([]ContentItem, error) {
	manifest, err := m.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	var filtered []ContentItem
	for _, item := range manifest.Items {
		var status PlatformStatus
		switch platform {
		case "youtube":
			status = item.Status.YouTube
		case "tiktok":
			status = item.Status.TikTok
		case "instagram":
			status = item.Status.Instagram
		case "twitter":
			status = item.Status.Twitter
		case "facebook":
			status = item.Status.Facebook
		case "linkedin":
			status = item.Status.LinkedIn
		default:
			return nil, fmt.Errorf("unknown platform: %s", platform)
		}

		if status.Posted == posted {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// DeleteItem removes an item from the manifest
func (m *ManifestManager) DeleteItem(id string) error {
	manifest, err := m.LoadManifest()
	if err != nil {
		return fmt.Errorf("failed to load manifest: %w", err)
	}

	// Find and remove item
	found := false
	for i, item := range manifest.Items {
		if item.ID == id {
			manifest.Items = append(manifest.Items[:i], manifest.Items[i+1:]...)
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("item with ID %s not found", id)
	}

	// Save
	return m.SaveManifest(*manifest)
}

// GetStats returns statistics about the manifest
func (m *ManifestManager) GetStats() (map[string]interface{}, error) {
	manifest, err := m.LoadManifest()
	if err != nil {
		return nil, fmt.Errorf("failed to load manifest: %w", err)
	}

	stats := map[string]interface{}{
		"total_items": len(manifest.Items),
		"platforms": map[string]interface{}{
			"youtube":   countPosted(manifest.Items, "youtube"),
			"tiktok":    countPosted(manifest.Items, "tiktok"),
			"instagram": countPosted(manifest.Items, "instagram"),
			"twitter":   countPosted(manifest.Items, "twitter"),
			"facebook":  countPosted(manifest.Items, "facebook"),
			"linkedin":  countPosted(manifest.Items, "linkedin"),
		},
	}

	return stats, nil
}

// Helper function to count posted items for a platform
func countPosted(items []ContentItem, platform string) int {
	count := 0
	for _, item := range items {
		var status PlatformStatus
		switch platform {
		case "youtube":
			status = item.Status.YouTube
		case "tiktok":
			status = item.Status.TikTok
		case "instagram":
			status = item.Status.Instagram
		case "twitter":
			status = item.Status.Twitter
		case "facebook":
			status = item.Status.Facebook
		case "linkedin":
			status = item.Status.LinkedIn
		}

		if status.Posted {
			count++
		}
	}
	return count
}

