package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/JAreyes98/healthconnect-storage-service/internal/crypto"
)

type LocalStrategy struct{}

type LocalConfig struct {
	BasePath string `json:"path"`
}

func (s *LocalStrategy) Upload(src io.Reader, filename string, configJSON string, shouldEncrypt bool) (string, error) {
	var cfg LocalConfig

	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("config error: %v", err)
	}

	if cfg.BasePath == "" {
		return "", fmt.Errorf("basePath is empty. Check if JSON key is 'path'. Raw config: %s", configJSON)
	}

	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return "", fmt.Errorf("mkdir failed for path [%s]: %v", cfg.BasePath, err)
	}

	fullPath := filepath.Join(cfg.BasePath, filename)

	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("read stream error: %v", err)
	}

	if shouldEncrypt {
		data, err = crypto.Encrypt(data)
		if err != nil {
			return "", fmt.Errorf("encryption error: %v", err)
		}
	}

	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("disk write error: %v", err)
	}

	return fullPath, nil
}

func (s *LocalStrategy) Download(path string, config string) (io.ReadCloser, error) {
	return os.Open(path)
}
