package storage

import "io"

type StorageStrategy interface {
	Upload(src io.Reader, filename string, config string, shouldEncrypt bool) (string, error)
	Download(path string, config string) (io.ReadCloser, error)
}

// Factory para obtener la estrategia según el tipo
func GetStrategy(pType string) StorageStrategy {
	switch pType {
	case "LOCAL":
		return &LocalStrategy{}
	case "S3":
		return &S3Strategy{}
	default:
		return &LocalStrategy{}
	}
}
