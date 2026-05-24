package store

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type S3Service struct {
	bucket        string
	client        *s3.Client
	presignClient *s3.PresignClient
	uploadExpiry  time.Duration
}

func NewS3Service(s3Client *s3.Client, bucketName string, uploadExpiry time.Duration) *S3Service {
	return &S3Service{
		bucket:        bucketName,
		client:        s3Client,
		presignClient: s3.NewPresignClient(s3Client),
		uploadExpiry:  uploadExpiry,
	}
}

var _ Store = (*S3Service)(nil)

// Generate a presigned upload url
func (s *S3Service) GenerateUploadURL(ctx context.Context, key string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	resp, err := s.presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = s.uploadExpiry
	})
	if err != nil {
		log.Println("Failed to presign upload object: ", err.Error())
		return "", fmt.Errorf("Internal Server Error")
	}

	return resp.URL, nil
}

// Check if the key exist
func (s *S3Service) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		// Not found
		var notFoundErr *types.NotFound
		if errors.As(err, &notFoundErr) {
			return false, nil
		}

		// Fallback
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" || apiErr.ErrorCode() == "NoSuchKey" {
				return false, nil
			}
		}

		log.Println("failed checking object existence: ", err.Error())
		return false, fmt.Errorf("Internal Server Error")
	}

	return true, nil
}

// Delete the item with the given key
func (s *S3Service) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}
