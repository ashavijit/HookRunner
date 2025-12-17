package policy

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewCache(t *testing.T) {
	c := NewCache("")
	if c == nil {
		t.Fatal("expected cache, got nil")
	}
}

func TestCache_KeyFor(t *testing.T) {
	c := NewCache("")
	k1 := c.KeyFor("http://a.com/x.yaml")
	k2 := c.KeyFor("http://b.com/x.yaml")

	if k1 == k2 {
		t.Error("different URLs should have different keys")
	}

	if len(k1) != 64 {
		t.Errorf("expected sha256 hex, got len %d", len(k1))
	}
}

func TestCache_MemoryGetSet(t *testing.T) {
	c := NewCache("")

	url := "http://example.com/test.yaml"
	policy := &RemotePolicy{Name: "test", Version: "1.0"}

	c.SetInMemory(url, policy)

	got := c.GetFromMemory(url)
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if got.Name != "test" {
		t.Errorf("got %s, want test", got.Name)
	}
}

func TestCache_DiskStorage(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	url := "https://example.com/test.yaml"
	policy := &RemotePolicy{Name: "disk-test", Version: "2.0"}
	data := []byte("name: disk-test\nversion: \"2.0\"\n")

	err := c.SaveToDisk(url, policy, data, "etag123")
	if err != nil {
		t.Fatalf("SaveToDisk failed: %v", err)
	}

	loaded, meta, err := c.GetFromDisk(url)
	if err != nil {
		t.Fatalf("GetFromDisk failed: %v", err)
	}

	if loaded.Name != "disk-test" {
		t.Errorf("got %s, want disk-test", loaded.Name)
	}
	if meta.ETag != "etag123" {
		t.Errorf("got etag %s, want etag123", meta.ETag)
	}
	if meta.URL != url {
		t.Errorf("got url %s, want %s", meta.URL, url)
	}
}

func TestCache_Clear(t *testing.T) {
	dir := t.TempDir()
	c := NewCache(dir)

	url := "https://example.com/clear.yaml"
	policy := &RemotePolicy{Name: "clear-test"}
	data := []byte("name: clear-test\n")

	if err := c.SaveToDisk(url, policy, data, ""); err != nil {
		t.Fatalf("SaveToDisk failed: %v", err)
	}
	c.SetInMemory(url, policy)

	err := c.Clear()
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	if c.GetFromMemory(url) != nil {
		t.Error("memory should be cleared")
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("disk cache should be deleted")
	}
}

func TestCacheDirCreation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "nested", "cache")
	c := NewCache(dir)

	url := "https://example.com/nested.yaml"
	policy := &RemotePolicy{Name: "nested"}
	data := []byte("name: nested\n")

	err := c.SaveToDisk(url, policy, data, "")
	if err != nil {
		t.Fatalf("SaveToDisk should create dir: %v", err)
	}
}
