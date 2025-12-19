package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type Cache struct {
	dir string
}

func New(workDir string) *Cache {
	return &Cache{
		dir: filepath.Join(workDir, ".hookrunner", "cache"),
	}
}

func (c *Cache) IsCached(hookName string, files []string, hookHash string) (cached, uncached []string) {
	for _, file := range files {
		fileHash, err := computeFileHash(file)
		if err != nil {
			uncached = append(uncached, file)
			continue
		}

		cacheKey := computeCacheKey(fileHash, hookHash)
		cachePath := c.getCachePath(hookName, cacheKey)

		if _, err := os.Stat(cachePath); err == nil {
			cached = append(cached, file)
		} else {
			uncached = append(uncached, file)
		}
	}
	return
}

func (c *Cache) MarkPassed(hookName string, files []string, hookHash string) error {
	hookDir := filepath.Join(c.dir, sanitizeName(hookName))
	if err := os.MkdirAll(hookDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		fileHash, err := computeFileHash(file)
		if err != nil {
			continue
		}

		cacheKey := computeCacheKey(fileHash, hookHash)
		cachePath := c.getCachePath(hookName, cacheKey)

		f, err := os.Create(cachePath)
		if err != nil {
			return err
		}
		f.Close()
	}
	return nil
}

func (c *Cache) Invalidate(hookName string, files []string, hookHash string) error {
	for _, file := range files {
		fileHash, err := computeFileHash(file)
		if err != nil {
			continue
		}

		cacheKey := computeCacheKey(fileHash, hookHash)
		cachePath := c.getCachePath(hookName, cacheKey)
		os.Remove(cachePath)
	}
	return nil
}

func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}

func (c *Cache) getCachePath(hookName, cacheKey string) string {
	return filepath.Join(c.dir, sanitizeName(hookName), cacheKey+".ok")
}

func computeFileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}

func computeCacheKey(fileHash, hookHash string) string {
	combined := fileHash + hookHash
	h := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(h[:])
}

func ComputeHookHash(tool string, args []string, files string, glob string, exclude string) string {
	data := tool + strings.Join(args, "|") + files + glob + exclude
	h := sha256.Sum256([]byte(data))
	return hex.EncodeToString(h[:])
}

func sanitizeName(name string) string {
	return strings.ReplaceAll(strings.ReplaceAll(name, "/", "_"), "\\", "_")
}
