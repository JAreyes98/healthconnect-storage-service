package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"github.com/JAreyes98/healthconnect-storage-service/internal/crypto"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox"
	"github.com/dropbox/dropbox-sdk-go-unofficial/v6/dropbox/files"
)

type DropboxStrategy struct{}

type DropboxConfig struct {
	AccessToken string `json:"access_token"`
	RootPath    string `json:"path"`
}

func (s *DropboxStrategy) Upload(src io.Reader, filename string, configJSON string, shouldEncrypt bool) (string, error) {
	var cfg DropboxConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("dropbox config error: %v", err)
	}

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

	dbxCfg := dropbox.Config{
		Token: cfg.AccessToken,
	}
	client := files.New(dbxCfg)

	fullPath := filepath.Join(cfg.RootPath, filename)
	if fullPath[0] != '/' {
		fullPath = "/" + fullPath
	}

	commitInfo := files.NewCommitInfo(fullPath)
	commitInfo.Mode = &files.WriteMode{Tagged: dropbox.Tagged{Tag: "overwrite"}}

	uploadArg := &files.UploadArg{
		CommitInfo: *commitInfo,
	}

	_, err = client.Upload(uploadArg, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("dropbox api upload error: %v", err)
	}

	return fullPath, nil
}

func (s *DropboxStrategy) Download(path string, configJSON string) (io.ReadCloser, error) {
	var cfg DropboxConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("dropbox config error: %v", err)
	}

	dbxCfg := dropbox.Config{
		Token: cfg.AccessToken,
	}
	client := files.New(dbxCfg)

	downloadArg := files.NewDownloadArg(path)
	_, content, err := client.Download(downloadArg)
	if err != nil {
		return nil, fmt.Errorf("dropbox api download error: %v", err)
	}

	return content, nil
}
