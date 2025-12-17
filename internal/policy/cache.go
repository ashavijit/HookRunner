package policy

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type CacheMetadata struct {
	URL         string    `json:"url"`
	ETag        string    `json:"etag"`
	LastFetched time.Time `json:"lastFetched"`
}

type Cache struct {
	dir    string
	memory map[string]*RemotePolicy
	mu     sync.RWMutex
}

func NewCache(dir string) *Cache {
	return &Cache{
		dir:    dir,
		memory: make(map[string]*RemotePolicy),
	}
}

func (c *Cache) KeyFor(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}

func (c *Cache) GetFromMemory(url string) *RemotePolicy {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.memory[c.KeyFor(url)]
}

func (c *Cache) SetInMemory(url string, policy *RemotePolicy) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.memory[c.KeyFor(url)] = policy
}

func (c *Cache) GetFromDisk(url string) (*RemotePolicy, *CacheMetadata, error) {
	key := c.KeyFor(url)
	cacheDir := filepath.Join(c.dir, "sha256_"+key)

	metaPath := filepath.Join(cacheDir, "metadata.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		return nil, nil, err
	}

	var meta CacheMetadata
	if err := json.Unmarshal(metaData, &meta); err != nil {
		return nil, nil, err
	}

	policyPath := filepath.Join(cacheDir, "policy.yaml")
	policyData, err := os.ReadFile(policyPath)
	if err != nil {
		return nil, &meta, err
	}

	policy, err := ParseRemotePolicy(policyData)
	if err != nil {
		return nil, &meta, err
	}

	return policy, &meta, nil
}

func (c *Cache) SaveToDisk(url string, policy *RemotePolicy, policyData []byte, etag string) error {
	key := c.KeyFor(url)
	cacheDir := filepath.Join(c.dir, "sha256_"+key)

	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	meta := CacheMetadata{
		URL:         url,
		ETag:        etag,
		LastFetched: time.Now(),
	}

	metaData, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return err
	}

	if err := os.WriteFile(filepath.Join(cacheDir, "metadata.json"), metaData, 0644); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(cacheDir, "policy.yaml"), policyData, 0644)
}

func (c *Cache) Clear() error {
	c.mu.Lock()
	c.memory = make(map[string]*RemotePolicy)
	c.mu.Unlock()

	if c.dir == "" {
		return nil
	}
	return os.RemoveAll(c.dir)
}
