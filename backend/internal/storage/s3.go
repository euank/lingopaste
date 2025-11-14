package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Storage struct {
	client     *s3.Client
	bucketName string
}

func NewS3Storage(ctx context.Context, region, bucketName string) (*S3Storage, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		client:     client,
		bucketName: bucketName,
	}, nil
}

func (s *S3Storage) SaveOriginal(ctx context.Context, pasteID, content string) error {
	key := fmt.Sprintf("pastes/%s/original.txt", pasteID)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(content),
		ContentType: aws.String("text/plain; charset=utf-8"),
	})
	return err
}

func (s *S3Storage) GetOriginal(ctx context.Context, pasteID string) (string, error) {
	key := fmt.Sprintf("pastes/%s/original.txt", pasteID)
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (s *S3Storage) SaveTranslation(ctx context.Context, pasteID, language, translation string) error {
	key := fmt.Sprintf("pastes/%s/translations/%s.txt", pasteID, language)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		Body:        strings.NewReader(translation),
		ContentType: aws.String("text/plain; charset=utf-8"),
	})
	return err
}

func (s *S3Storage) GetTranslation(ctx context.Context, pasteID, language string) (string, error) {
	key := fmt.Sprintf("pastes/%s/translations/%s.txt", pasteID, language)
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", err
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
