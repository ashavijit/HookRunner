package tool

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ashavijit/hookrunner/internal/config"
)

type Manager struct {
	CacheDir string
}

func NewManager(cacheDir string) *Manager {
	return &Manager{CacheDir: cacheDir}
}

func (m *Manager) EnsureTool(name string, tool *config.Tool) (string, error) {
	if tool == nil {
		return m.findSystemTool(name)
	}

	cachedPath := m.getCachedPath(name, tool.Version)
	if _, err := os.Stat(cachedPath); err == nil {
		return cachedPath, nil
	}

	url := tool.Install[runtime.GOOS]
	if url == "" {
		return "", fmt.Errorf("no download URL for %s on %s", name, runtime.GOOS)
	}

	if err := m.downloadAndExtract(name, tool.Version, url, tool.Checksum); err != nil {
		return "", err
	}

	return cachedPath, nil
}

func (m *Manager) findSystemTool(name string) (string, error) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", fmt.Errorf("tool %s not found in PATH", name)
	}
	return path, nil
}

func (m *Manager) getCachedPath(name, version string) string {
	binName := name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	return filepath.Join(m.CacheDir, fmt.Sprintf("%s-%s", name, version), binName)
}

func (m *Manager) downloadAndExtract(name, version, url, checksum string) error {
	//nolint:gosec // G107: URL is from trusted config file
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("download failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: HTTP %d", resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp("", "hookrunner-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	hasher := sha256.New()
	writer := io.MultiWriter(tmpFile, hasher)
	if _, err := io.Copy(writer, resp.Body); err != nil {
		return fmt.Errorf("download failed: %w", err)
	}

	if checksum != "" {
		got := hex.EncodeToString(hasher.Sum(nil))
		if got != checksum {
			return fmt.Errorf("checksum mismatch: expected %s, got %s", checksum, got)
		}
	}

	destDir := filepath.Join(m.CacheDir, fmt.Sprintf("%s-%s", name, version))
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache dir: %w", err)
	}

	if strings.HasSuffix(url, ".tar.gz") || strings.HasSuffix(url, ".tgz") {
		return m.extractTarGz(tmpFile.Name(), destDir, name)
	} else if strings.HasSuffix(url, ".zip") {
		return m.extractZip(tmpFile.Name(), destDir, name)
	}

	return fmt.Errorf("unsupported archive format")
}

func (m *Manager) extractTarGz(src, dest, toolName string) error {
	file, err := os.Open(src)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header.Typeflag != tar.TypeReg {
			continue
		}

		baseName := filepath.Base(header.Name)
		if strings.HasPrefix(baseName, toolName) {
			destPath := filepath.Join(dest, baseName)
			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				return err
			}
			//nolint:gosec // G110: Trusted archive from configured URLs
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
}

func (m *Manager) extractZip(src, dest, toolName string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}

		baseName := filepath.Base(f.Name)
		if strings.HasPrefix(baseName, toolName) {
			destPath := filepath.Join(dest, baseName)
			rc, err := f.Open()
			if err != nil {
				return err
			}

			outFile, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY, 0755)
			if err != nil {
				rc.Close()
				return err
			}

			//nolint:gosec // G110: Trusted archive from configured URLs
			_, err = io.Copy(outFile, rc)
			outFile.Close()
			rc.Close()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
