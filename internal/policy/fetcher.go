package policy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Fetcher struct {
	client *http.Client
	cache  *Cache
}

func NewFetcher(cacheDir string) *Fetcher {
	return &Fetcher{
		client: &http.Client{Timeout: 30 * time.Second},
		cache:  NewCache(filepath.Join(cacheDir, "policies")),
	}
}

func (f *Fetcher) LoadPolicy(url string) (*RemotePolicy, error) {
	if err := f.validateURL(url); err != nil {
		return nil, err
	}

	if policy := f.cache.GetFromMemory(url); policy != nil {
		return policy, nil
	}

	cached, meta, diskErr := f.cache.GetFromDisk(url)
	_ = diskErr // Intentionally ignore - fallback to network if cache fails
	if cached != nil && meta != nil {
		notModified, err := f.checkNotModified(url, meta.ETag)
		if err == nil && notModified {
			f.cache.SetInMemory(url, cached)
			return cached, nil
		}
	}

	policy, data, etag, err := f.fetchFromNetwork(url)
	if err != nil {
		if cached != nil {
			f.cache.SetInMemory(url, cached)
			return cached, nil
		}
		return nil, err
	}

	if err := ValidatePolicy(policy); err != nil {
		return nil, fmt.Errorf("invalid policy: %w", err)
	}

	_ = f.cache.SaveToDisk(url, policy, data, etag) // Best-effort cache
	f.cache.SetInMemory(url, policy)

	return policy, nil
}

func (f *Fetcher) validateURL(url string) error {
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("HTTPS required: %s", url)
	}
	return nil
}

func (f *Fetcher) checkNotModified(url, etag string) (bool, error) {
	if etag == "" {
		return false, nil
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("If-None-Match", etag)

	resp, err := f.client.Do(req)
	if err != nil {
		return false, err
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusNotModified, nil
}

func (f *Fetcher) fetchFromNetwork(url string) (*RemotePolicy, []byte, string, error) {
	resp, err := f.client.Get(url)
	if err != nil {
		return nil, nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, "", err
	}

	policy, err := ParseRemotePolicy(data)
	if err != nil {
		return nil, nil, "", err
	}

	return policy, data, resp.Header.Get("ETag"), nil
}

func (f *Fetcher) ClearCache() error {
	return f.cache.Clear()
}

func ParseRemotePolicy(data []byte) (*RemotePolicy, error) {
	var policy RemotePolicy

	if err := yaml.Unmarshal(data, &policy); err != nil {
		if err := json.Unmarshal(data, &policy); err != nil {
			return nil, fmt.Errorf("invalid policy format")
		}
	}

	return &policy, nil
}
