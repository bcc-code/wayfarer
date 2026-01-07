package services

import (
	"context"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	appconfig "github.com/bcc-media/wayfarer/internal/config"
)

// S3Service handles file uploads to AWS S3
type S3Service struct {
	client        *s3.Client
	bucket        string
	region        string
	publicBaseURL string
}

// NewS3Service creates a new S3 service with the given configuration.
// Uses static AWS credentials from configuration.
func NewS3Service(cfg appconfig.S3Config) (*S3Service, error) {
	awsCfg := aws.Config{
		Region: cfg.Region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Service{
		client:        client,
		bucket:        cfg.Bucket,
		region:        cfg.Region,
		publicBaseURL: cfg.PublicBaseURL,
	}, nil
}

// UploadFile uploads a file to S3 and returns the public URL
func (s *S3Service) UploadFile(ctx context.Context, file io.Reader, filename string, contentType string, fileSize int64) (string, error) {
	// Upload to S3
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(filename),
		Body:          file,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(fileSize),
		ACL:           types.ObjectCannedACLPublicRead,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct public URL
	var publicURL string
	if s.publicBaseURL != "" {
		publicURL = fmt.Sprintf("%s/%s", s.publicBaseURL, filename)
	} else {
		publicURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, filename)
	}

	return publicURL, nil
}
