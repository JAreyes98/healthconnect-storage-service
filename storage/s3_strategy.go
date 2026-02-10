package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Config struct {
	BucketName string `json:"bucket_name"`
	Region     string `json:"region"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
	RootFolder string `json:"path"`
}

type S3Strategy struct{}

func (s *S3Strategy) Upload(src io.Reader, filename string, configJSON string, shouldEncrypt bool) (string, error) {
	var cfg S3Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return "", err
	}

	staticResolver := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(staticResolver),
	)
	if err != nil {
		return "", err
	}

	client := s3.NewFromConfig(sdkConfig)

	fullKey := path.Join(cfg.RootFolder, filename)
	fullKey = strings.TrimPrefix(fullKey, "/")

	data, err := io.ReadAll(src)
	if err != nil {
		return "", err
	}

	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(fullKey),
		Body:   bytes.NewReader(data),
	})

	if err != nil {
		return "", fmt.Errorf("s3 upload error: %w", err)
	}

	return fullKey, nil
}

func (s *S3Strategy) Download(filePath string, configJSON string) (io.ReadCloser, error) {
	var cfg S3Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, err
	}

	staticResolver := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(staticResolver),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(sdkConfig)

	finalKey := filePath
	if !strings.HasPrefix(strings.TrimPrefix(filePath, "/"), strings.TrimPrefix(cfg.RootFolder, "/")) {
		finalKey = path.Join(cfg.RootFolder, filePath)
	}
	finalKey = strings.TrimPrefix(finalKey, "/")

	result, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(finalKey),
	})

	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (s *S3Strategy) Delete(filePath string, configJSON string) error {
	var cfg S3Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return err
	}

	staticResolver := credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")
	sdkConfig, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(staticResolver),
	)
	if err != nil {
		return err
	}

	client := s3.NewFromConfig(sdkConfig)

	finalKey := strings.TrimPrefix(filePath, "/")

	_, err = client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(cfg.BucketName),
		Key:    aws.String(finalKey),
	})

	return err
}
