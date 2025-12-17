package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type HookResult struct {
	Success   bool      `json:"success"`
	Output    string    `json:"output"`
	Timestamp time.Time `json:"timestamp"`
}

type FileCache struct {
	dir string
}

func New(cacheDir string) *FileCache {
	dir := filepath.Join(cacheDir, "results")
	os.MkdirAll(dir, 0755)
	return &FileCache{dir: dir}
}

func (c *FileCache) cacheKey(hookName string, files []string) string {
	h := sha256.New()
	h.Write([]byte(hookName))

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}
		h.Write([]byte(file))
		h.Write(data)
	}

	return hex.EncodeToString(h.Sum(nil))[:16]
}

func (c *FileCache) Get(hookName string, files []string) (*HookResult, bool) {
	key := c.cacheKey(hookName, files)
	path := filepath.Join(c.dir, key+".json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var result HookResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, false
	}

	if time.Since(result.Timestamp) > 24*time.Hour {
		os.Remove(path)
		return nil, false
	}

	return &result, true
}

func (c *FileCache) Set(hookName string, files []string, success bool, output string) {
	key := c.cacheKey(hookName, files)
	path := filepath.Join(c.dir, key+".json")

	result := HookResult{
		Success:   success,
		Output:    output,
		Timestamp: time.Now(),
	}

	data, _ := json.Marshal(result)
	os.WriteFile(path, data, 0644)
}

func (c *FileCache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		os.Remove(filepath.Join(c.dir, e.Name()))
	}
	return nil
}

func (c *FileCache) Stats() (int, int64) {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return 0, 0
	}

	var size int64
	for _, e := range entries {
		info, _ := e.Info()
		if info != nil {
			size += info.Size()
		}
	}

	return len(entries), size
}

func FormatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
}
