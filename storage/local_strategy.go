package storage

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/JAreyes98/healthconnect-storage-service/internal/crypto"
)

type LocalStrategy struct{}

type LocalConfig struct {
	BasePath string `json:"basePath"`
}

func (s *LocalStrategy) Upload(src io.Reader, filename string, configJSON string, shouldEncrypt bool) (string, error) {
	var cfg LocalConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("error al leer config: %v", err)
	}

	if err := os.MkdirAll(cfg.BasePath, 0755); err != nil {
		return "", err
	}

	fullPath := filepath.Join(cfg.BasePath, filename)

	// 1. LEER TODO EL CONTENIDO PRIMERO
	data, err := io.ReadAll(src)
	if err != nil {
		return "", fmt.Errorf("error leyendo stream: %v", err)
	}

	// 2. CIFRAR SI ES NECESARIO
	if shouldEncrypt {
		// Asumiendo que creaste el paquete internal/crypto
		data, err = crypto.Encrypt(data)
		if err != nil {
			return "", fmt.Errorf("error cifrando: %v", err)
		}
		log.Println("🔐 Archivo cifrado exitosamente")
	}

	// 3. ESCRIBIR EL RESULTADO FINAL AL DISCO
	// os.WriteFile crea el archivo y escribe los bytes de un solo golpe
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("error guardando archivo: %v", err)
	}

	return fullPath, nil
}

func (s *LocalStrategy) Download(path string, config string) (io.ReadCloser, error) {
	return os.Open(path)
}
