package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

const (
	MaxImageSize = 10 * 1024 * 1024  // 10MB
	MaxVideoSize = 100 * 1024 * 1024 // 100MB
	MaxPDFSize   = 10 * 1024 * 1024  // 10MB
	MaxXLSXSize  = 10 * 1024 * 1024  // 10MB
)

var (
	Bucket                *GCPBucket
	AllowedFileExtensions = []string{"png", "jpg", "jpeg", "mp4", "pdf", "xlsx"}
)

type GCPBucket struct {
	client         *storage.Client
	bucketName     string
	credPath       string
	serviceAccount *ServiceAccount
}

type ServiceAccount struct {
	ClientEmail string `json:"client_email"`
	PrivateKey  string `json:"private_key"`
}

type UploadResult struct {
	Filename   string `json:"filename"`
	Size       int64  `json:"size"`
	ObjectName string `json:"object_name"`
}

func GCPBucketApp(credPath *string) error {
	bucket, err := NewGCPBucket(credPath)
	if err != nil {
		return err
	}
	Bucket = bucket
	return nil
}

func NewGCPBucket(credPath *string) (*GCPBucket, error) {
	if credPath == nil {
		defaultPath := filepath.Join("internal", "config", "creds", "gcp_bucket.json")
		credPath = &defaultPath
	}

	bucket := &GCPBucket{
		credPath:   *credPath,
		bucketName: AppConfig.GCPBucketName,
	}

	if err := bucket.loadServiceAccount(); err != nil {
		return nil, fmt.Errorf("failed to load service account: %w", err)
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(*credPath))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Google Cloud Storage: %w", err)
	}

	bucket.client = client
	return bucket, nil
}

func (b *GCPBucket) Client() *storage.Client {
	return b.client
}

func (b *GCPBucket) BucketName() string {
	return b.bucketName
}

func (b *GCPBucket) CredPath() string {
	return b.credPath
}

func (b *GCPBucket) ServiceAccountEmail() string {
	return b.serviceAccount.ClientEmail
}

func (b *GCPBucket) ServiceAccountPrivateKey() []byte {
	return []byte(b.serviceAccount.PrivateKey)
}

func (b *GCPBucket) loadServiceAccount() error {
	data, err := os.ReadFile(b.credPath)
	if err != nil {
		return fmt.Errorf("failed to read service account file: %w", err)
	}

	var sa ServiceAccount
	if err := json.Unmarshal(data, &sa); err != nil {
		return fmt.Errorf("failed to parse service account file: %w", err)
	}

	b.serviceAccount = &sa
	return nil
}

func (b *GCPBucket) Close() error {
	if b.client != nil {
		return b.client.Close()
	}
	return nil
}

func (b *GCPBucket) GetFile(ctx context.Context, name string) (io.ReadCloser, string, error) {
	obj := b.client.Bucket(b.bucketName).Object(name)

	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("file not found: %w", err)
	}

	attrs, err := obj.Attrs(ctx)
	contentType := ""
	if err == nil {
		contentType = attrs.ContentType
	}

	return rc, contentType, nil
}

func (b *GCPBucket) Upload(ctx context.Context, file multipart.File, header *multipart.FileHeader) (*UploadResult, error) {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(header.Filename), "."))

	if !b.isAllowedExtension(ext) {
		return nil, fmt.Errorf("file type not allowed: %s", header.Filename)
	}

	if err := b.validateFileSize(ext, header.Size); err != nil {
		return nil, err
	}

	bucket := b.client.Bucket(b.bucketName)
	objectName := fmt.Sprintf("%d_%s", time.Now().UnixNano(), header.Filename)
	wc := bucket.Object(objectName).NewWriter(ctx)

	if _, err := io.Copy(wc, file); err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	if err := wc.Close(); err != nil {
		return nil, fmt.Errorf("failed to close writer: %w", err)
	}

	return &UploadResult{
		Filename:   header.Filename,
		Size:       header.Size,
		ObjectName: objectName,
	}, nil
}

func (b *GCPBucket) Delete(ctx context.Context, name string) error {
	obj := b.client.Bucket(b.bucketName).Object(name)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (b *GCPBucket) GenerateSignedURL(object string, expiry time.Duration) (string, error) {
	url, err := storage.SignedURL(b.bucketName, object, &storage.SignedURLOptions{
		GoogleAccessID: b.serviceAccount.ClientEmail,
		PrivateKey:     []byte(b.serviceAccount.PrivateKey),
		Method:         "GET",
		Expires:        time.Now().Add(expiry),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}
	return url, nil
}

func (b *GCPBucket) isAllowedExtension(ext string) bool {
	for _, e := range AllowedFileExtensions {
		if e == ext {
			return true
		}
	}
	return false
}

func (b *GCPBucket) validateFileSize(ext string, size int64) error {
	switch ext {
	case "png", "jpg", "jpeg":
		if size > MaxImageSize {
			return fmt.Errorf("image too large (max %dMB)", MaxImageSize/1024/1024)
		}
	case "mp4":
		if size > MaxVideoSize {
			return fmt.Errorf("video too large (max %dMB)", MaxVideoSize/1024/1024)
		}
	case "pdf":
		if size > MaxPDFSize {
			return fmt.Errorf("PDF too large (max %dMB)", MaxPDFSize/1024/1024)
		}
	case "xlsx":
		if size > MaxXLSXSize {
			return fmt.Errorf("XLSX too large (max %dMB)", MaxXLSXSize/1024/1024)
		}
	}
	return nil
}
