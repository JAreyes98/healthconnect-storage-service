package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/JAreyes98/healthconnect-storage-service/internal/crypto"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Strategy struct{}

type S3Config struct {
	Region     string `json:"region"`
	BucketName string `json:"bucket_name"`
}

func (s *S3Strategy) Upload(src io.Reader, filename string, configJSON string, shouldEncrypt bool) (string, error) {
	var cfg S3Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid S3 config: %v", err)
	}

	// 1. Leer y procesar datos (Cifrado)
	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	if shouldEncrypt {
		data, err = crypto.Encrypt(data)
		if err != nil {
			return "", err
		}
	}

	// 2. Configurar cliente S3
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.Region))
	if err != nil {
		return "", fmt.Errorf("unable to load SDK config: %v", err)
	}
	client := s3.NewFromConfig(awsCfg)

	// 3. Subir a S3
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(filename),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %v", err)
	}

	// Retornamos el nombre del objeto como path físico
	return filename, nil
}

// Download retrieves a file from S3 and returns a reader
func (s *S3Strategy) Download(path string, configJSON string) (io.ReadCloser, error) {
	var cfg S3Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid S3 config: %v", err)
	}

	// 1. Configurar cliente AWS
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config: %v", err)
	}
	client := s3.NewFromConfig(awsCfg)

	// 2. Obtener objeto de S3
	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(path),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object from S3: %v", err)
	}

	// Retornamos el Body (que es un io.ReadCloser)
	return result.Body, nil
}
