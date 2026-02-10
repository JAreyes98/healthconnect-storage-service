package storage

import "io"

type StorageStrategy interface {
	Upload(src io.Reader, filename string, config string, shouldEncrypt bool) (string, error)
	Download(path string, config string) (io.ReadCloser, error)
}

// Factory para obtener la estrategia según el tipo
func GetStrategy(pType string) (StorageStrategy, bool) {
	switch pType {
	case "LOCAL":
		return &LocalStrategy{}, true
	case "S3":
		return &S3Strategy{}, true
	case "DROPBOX":
		return &DropboxStrategy{}, true
	default:
		return nil, false
	}
}
