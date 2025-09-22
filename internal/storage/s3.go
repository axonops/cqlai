package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Backend implements Backend for Amazon S3 and S3-compatible storage
type S3Backend struct {
	config   *Config
	client   *s3.Client
	uploader *manager.Uploader
	bucket   string
}

// NewS3Backend creates a new S3 backend
func NewS3Backend(cfg *Config) (*S3Backend, error) {
	// Configure AWS SDK
	awsConfig, err := buildAWSConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to build AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsConfig, func(o *s3.Options) {
		// For S3-compatible services
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.PathStyle
		}
	})

	// Create uploader for efficient streaming uploads
	uploader := manager.NewUploader(client, func(u *manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024 // 5MB parts
		u.Concurrency = 3
	})

	return &S3Backend{
		config:   cfg,
		client:   client,
		uploader: uploader,
		bucket:   cfg.Bucket,
	}, nil
}

// buildAWSConfig builds AWS configuration from our config
func buildAWSConfig(cfg *Config) (aws.Config, error) {
	var opts []func(*config.LoadOptions) error

	// Set region
	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	// Set credentials if provided
	if cfg.AccessKeyID != "" && cfg.SecretAccessKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				cfg.SessionToken,
			),
		))
	}

	// Load config
	awsConfig, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return awsConfig, err
	}

	return awsConfig, nil
}

// s3Writer implements Writer for S3
type s3Writer struct {
	pipeReader *io.PipeReader
	pipeWriter *io.PipeWriter
	uploadErr  chan error
	url        string
	done       chan struct{}
}

func (w *s3Writer) Write(p []byte) (n int, err error) {
	return w.pipeWriter.Write(p)
}

func (w *s3Writer) Close() error {
	// Close the pipe writer to signal end of data
	err := w.pipeWriter.Close()

	// Wait for upload to complete
	<-w.done

	// Check for upload error
	select {
	case uploadErr := <-w.uploadErr:
		if uploadErr != nil {
			return uploadErr
		}
	default:
	}

	return err
}

func (w *s3Writer) URL() string {
	return w.url
}

// NewWriter creates a new writer for S3
func (b *S3Backend) NewWriter(ctx context.Context, path string) (Writer, error) {
	// Clean path - remove leading slash if present
	path = strings.TrimPrefix(path, "/")

	// Create pipe for streaming upload
	pr, pw := io.Pipe()

	// Prepare upload error channel and done signal
	uploadErr := make(chan error, 1)
	done := make(chan struct{})

	// Build URL
	url := fmt.Sprintf("s3://%s/%s", b.bucket, path)
	if b.config.Endpoint != "" {
		// For S3-compatible services, include the endpoint
		url = fmt.Sprintf("%s/%s/%s", b.config.Endpoint, b.bucket, path)
	}

	writer := &s3Writer{
		pipeReader: pr,
		pipeWriter: pw,
		uploadErr:  uploadErr,
		url:        url,
		done:       done,
	}

	// Start upload in background
	go func() {
		defer close(done)

		input := &s3.PutObjectInput{
			Bucket:      aws.String(b.bucket),
			Key:         aws.String(path),
			Body:        pr,
			ContentType: aws.String("application/octet-stream"),
		}

		// Add metadata if provided
		if b.config.Metadata != nil {
			input.Metadata = b.config.Metadata
		}

		_, err := b.uploader.Upload(ctx, input)
		if err != nil {
			uploadErr <- fmt.Errorf("S3 upload failed: %w", err)
			_ = pr.Close() // Ensure reader is closed on error
		}
	}()

	return writer, nil
}

// s3Reader implements Reader for S3
type s3Reader struct {
	body io.ReadCloser
	url  string
}

func (r *s3Reader) Read(p []byte) (n int, err error) {
	return r.body.Read(p)
}

func (r *s3Reader) Close() error {
	return r.body.Close()
}

func (r *s3Reader) URL() string {
	return r.url
}

// NewReader creates a new reader for S3
func (b *S3Backend) NewReader(ctx context.Context, path string) (Reader, error) {
	// Clean path
	path = strings.TrimPrefix(path, "/")

	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path),
	}

	result, err := b.client.GetObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get S3 object: %w", err)
	}

	url := fmt.Sprintf("s3://%s/%s", b.bucket, path)
	if b.config.Endpoint != "" {
		url = fmt.Sprintf("%s/%s/%s", b.config.Endpoint, b.bucket, path)
	}

	return &s3Reader{
		body: result.Body,
		url:  url,
	}, nil
}

// Exists checks if an object exists in S3
func (b *S3Backend) Exists(ctx context.Context, path string) (bool, error) {
	path = strings.TrimPrefix(path, "/")

	_, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		// Check if it's a not found error
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

// List lists objects with the given prefix
func (b *S3Backend) List(ctx context.Context, prefix string) ([]string, error) {
	prefix = strings.TrimPrefix(prefix, "/")

	var files []string

	paginator := s3.NewListObjectsV2Paginator(b.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(b.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list S3 objects: %w", err)
		}

		for _, obj := range output.Contents {
			if obj.Key != nil {
				files = append(files, *obj.Key)
			}
		}
	}

	return files, nil
}

// Delete removes an object from S3
func (b *S3Backend) Delete(ctx context.Context, path string) error {
	path = strings.TrimPrefix(path, "/")

	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(path),
	})

	if err != nil {
		return fmt.Errorf("failed to delete S3 object: %w", err)
	}

	return nil
}

// Provider returns the provider type
func (b *S3Backend) Provider() Provider {
	return ProviderS3
}