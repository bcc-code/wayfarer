package services

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"cloud.google.com/go/compute/metadata"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	appconfig "github.com/bcc-media/wayfarer/internal/config"
)

// S3Service handles file uploads to AWS S3
type S3Service struct {
	client        *s3.Client
	bucket        string
	region        string
	publicBaseURL string
}

// gcpTokenFetcher fetches OIDC tokens from Cloud Run metadata server
type gcpTokenFetcher struct {
	audience string
}

func (t gcpTokenFetcher) GetIdentityToken() ([]byte, error) {
	token, err := metadata.Get(
		fmt.Sprintf("instance/service-accounts/default/identity?audience=%s", t.audience),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get identity token from metadata server: %w", err)
	}

	// Debug: decode and log JWT claims
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payload, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var claims map[string]interface{}
			if json.Unmarshal(payload, &claims) == nil {
				slog.Info("OIDC token claims", "aud", claims["aud"], "iss", claims["iss"], "sub", claims["sub"], "email", claims["email"])
			}
		}
	}

	return []byte(token), nil
}

// NewS3Service creates a new S3 service with the given configuration.
// On Cloud Run with RoleARN set, it uses OIDC to assume the AWS role.
// Locally, it uses the default AWS credential chain (env vars, ~/.aws/credentials).
func NewS3Service(ctx context.Context, cfg appconfig.S3Config) (*S3Service, error) {
	// Load default AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// If on GCE/Cloud Run and RoleARN is set, use OIDC authentication
	if metadata.OnGCE() && cfg.RoleARN != "" {
		stsClient := sts.NewFromConfig(awsCfg)
		creds := stscreds.NewWebIdentityRoleProvider(
			stsClient,
			cfg.RoleARN,
			gcpTokenFetcher{audience: cfg.RoleARN},
		)
		awsCfg.Credentials = aws.NewCredentialsCache(creds)
	}
	// Otherwise, LoadDefaultConfig uses env vars or ~/.aws/credentials

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
