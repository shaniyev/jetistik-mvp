package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps MinIO operations.
type Client struct {
	mc     *minio.Client
	bucket string
}

// NewClient creates a new MinIO storage client and ensures the bucket exists.
func NewClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Client, error) {
	mc, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := mc.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := mc.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
		}
		slog.Info("created MinIO bucket", "bucket", bucket)
	}

	return &Client{mc: mc, bucket: bucket}, nil
}

// Upload stores a file in MinIO and returns the object path.
func (c *Client) Upload(ctx context.Context, objectPath string, reader io.Reader, size int64, contentType string) (string, error) {
	_, err := c.mc.PutObject(ctx, c.bucket, objectPath, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload %s: %w", objectPath, err)
	}
	return objectPath, nil
}

// Download returns a reader for the object at the given path.
func (c *Client) Download(ctx context.Context, objectPath string) (io.ReadCloser, error) {
	obj, err := c.mc.GetObject(ctx, c.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("download %s: %w", objectPath, err)
	}
	return obj, nil
}

// Delete removes an object from MinIO.
func (c *Client) Delete(ctx context.Context, objectPath string) error {
	err := c.mc.RemoveObject(ctx, c.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("delete %s: %w", objectPath, err)
	}
	return nil
}

// PresignedURL generates a presigned download URL valid for the given duration.
func (c *Client) PresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	url, err := c.mc.PresignedGetObject(ctx, c.bucket, objectPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presigned url %s: %w", objectPath, err)
	}
	return url.String(), nil
}

// TemplatePath returns the MinIO object path for a template file.
func TemplatePath(eventID int64, filename string) string {
	return fmt.Sprintf("templates/%d/%s", eventID, filename)
}

// ImportPath returns the MinIO object path for an import file.
func ImportPath(eventID int64, filename string) string {
	return fmt.Sprintf("imports/%d/%s", eventID, filename)
}

// CertificatePath returns the MinIO object path for a certificate PDF.
func CertificatePath(eventID int64, code string) string {
	return fmt.Sprintf("certificates/%d/%s.pdf", eventID, code)
}

// Ext returns the file extension from a filename.
func Ext(filename string) string {
	return path.Ext(filename)
}
