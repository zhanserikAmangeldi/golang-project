package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	UseSSL    bool
}

type MinioService struct {
	client *minio.Client
}

func NewMinioService(cfg MinioConfig) (*MinioService, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	log.Printf("MinIO client initialized: %s", cfg.Endpoint)

	service := &MinioService{client: client}
	if err := service.initBuckets(context.Background()); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *MinioService) initBuckets(ctx context.Context) error {
	buckets := []string{
		"chat-images",
		"chat-files",
		"chat-audio",
		"chat-video",
	}

	for _, bucket := range buckets {
		exists, err := s.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("failed to check bucket %s: %w", bucket, err)
		}

		if !exists {
			err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("failed to create bucket %s: %w", bucket, err)
			}
			log.Printf("Created MinIO bucket: %s", bucket)

			policy := fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {"AWS": ["*"]},
						"Action": ["s3:GetObject"],
						"Resource": ["arn:aws:s3:::%s/*"]
					}
				]
			}`, bucket)

			err = s.client.SetBucketPolicy(ctx, bucket, policy)
			if err != nil {
				log.Printf("Warning: failed to set policy for bucket %s: %v", bucket, err)
			}
		}
	}

	return nil
}

func (s *MinioService) UploadFile(ctx context.Context, bucket, objectName string, reader interface{}, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, bucket, objectName, reader.(interface {
		Read([]byte) (int, error)
		Seek(int64, int) (int64, error)
	}), size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinioService) GetFileURL(ctx context.Context, bucket, objectName string, expires time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, objectName, expires, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}

func (s *MinioService) DeleteFile(ctx context.Context, bucket, objectName string) error {
	return s.client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
}

func (s *MinioService) GetFileInfo(ctx context.Context, bucket, objectName string) (*minio.ObjectInfo, error) {
	info, err := s.client.StatObject(ctx, bucket, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func DetermineBucket(mimeType string) string {
	switch {
	case isImage(mimeType):
		return "chat-images"
	case isAudio(mimeType):
		return "chat-audio"
	case isVideo(mimeType):
		return "chat-video"
	default:
		return "chat-files"
	}
}

func isImage(mimeType string) bool {
	images := []string{
		"image/jpeg", "image/jpg", "image/png", "image/gif",
		"image/webp", "image/svg+xml", "image/bmp",
	}
	for _, img := range images {
		if mimeType == img {
			return true
		}
	}
	return false
}

func isAudio(mimeType string) bool {
	audios := []string{
		"audio/mpeg", "audio/mp3", "audio/wav", "audio/ogg",
		"audio/webm", "audio/aac", "audio/m4a",
	}
	for _, audio := range audios {
		if mimeType == audio {
			return true
		}
	}
	return false
}

func isVideo(mimeType string) bool {
	videos := []string{
		"video/mp4", "video/mpeg", "video/quicktime",
		"video/x-msvideo", "video/webm", "video/ogg",
	}
	for _, video := range videos {
		if mimeType == video {
			return true
		}
	}
	return false
}
