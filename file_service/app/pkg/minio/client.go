package minio

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	miniosdk "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"file_service/pkg/logging"
)

type ObjectInfo struct {
	Key         string
	Name        string
	UserUUID    string
	Size        int64
	ContentType string
}

type Client struct {
	logger      logging.Logger
	minioClient *miniosdk.Client
}

func NewClient(endpoint, accessKeyID, secretAccessKey string, useSSL bool, logger logging.Logger) (*Client, error) {
	if strings.TrimSpace(endpoint) == "" {
		return nil, errors.New("minio endpoint is required")
	}
	if strings.TrimSpace(accessKeyID) == "" || strings.TrimSpace(secretAccessKey) == "" {
		return nil, errors.New("minio credentials are required")
	}

	minioClient, err := miniosdk.New(endpoint, &miniosdk.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}

	return &Client{
		logger:      logger,
		minioClient: minioClient,
	}, nil
}

func (c *Client) EnsureBucket(ctx context.Context, bucketName string) (err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 10*time.Second)
	defer cancel()

	exists, err := c.minioClient.BucketExists(reqCtx, bucketName)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if exists {
		return nil
	}

	c.logger.Info("creating note bucket", "bucket", bucketName)
	if err = c.minioClient.MakeBucket(reqCtx, bucketName, miniosdk.MakeBucketOptions{}); err != nil {
		return fmt.Errorf("create bucket: %w", err)
	}

	return nil
}

func (c *Client) BucketExists(ctx context.Context, bucketName string) (exists bool, err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 10*time.Second)
	defer cancel()

	exists, err = c.minioClient.BucketExists(reqCtx, bucketName)
	if err != nil {
		return false, fmt.Errorf("check bucket exists: %w", err)
	}

	return exists, nil
}

func (c *Client) ListBuckets(ctx context.Context) (buckets []miniosdk.BucketInfo, err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 10*time.Second)
	defer cancel()

	buckets, err = c.minioClient.ListBuckets(reqCtx)
	if err != nil {
		return nil, fmt.Errorf("list buckets: %w", err)
	}

	return buckets, nil
}

func (c *Client) GetFile(ctx context.Context, bucketName, fileID string) (object *miniosdk.Object, err error) {
	object, err = c.minioClient.GetObject(resolveContext(ctx), bucketName, fileID, miniosdk.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("get file %s from bucket %s: %w", fileID, bucketName, err)
	}

	return object, nil
}

func (c *Client) StatFile(ctx context.Context, bucketName, fileID string) (info miniosdk.ObjectInfo, err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 10*time.Second)
	defer cancel()

	info, err = c.minioClient.StatObject(reqCtx, bucketName, fileID, miniosdk.StatObjectOptions{})
	if err != nil {
		return miniosdk.ObjectInfo{}, fmt.Errorf("stat file %s from bucket %s: %w", fileID, bucketName, err)
	}

	return info, nil
}

func (c *Client) ListFiles(ctx context.Context, bucketName string) (files []ObjectInfo, err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 15*time.Second)
	defer cancel()

	var result []ObjectInfo
	for object := range c.minioClient.ListObjects(reqCtx, bucketName, miniosdk.ListObjectsOptions{}) {
		if object.Err != nil {
			return nil, fmt.Errorf("list objects from bucket %s: %w", bucketName, object.Err)
		}

		info, statErr := c.minioClient.StatObject(reqCtx, bucketName, object.Key, miniosdk.StatObjectOptions{})
		if statErr != nil {
			return nil, fmt.Errorf("stat object %s from bucket %s: %w", object.Key, bucketName, statErr)
		}

		result = append(result, ObjectInfo{
			Key:         info.Key,
			Name:        objectName(info),
			UserUUID:    objectUserUUID(info),
			Size:        info.Size,
			ContentType: objectContentType(info),
		})
	}

	return result, nil
}

func (c *Client) UploadFile(ctx context.Context, bucketName, fileID, fileName, userUUID, contentType string, fileSize int64, reader io.Reader) (err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 30*time.Second)
	defer cancel()

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err = c.minioClient.PutObject(reqCtx, bucketName, fileID, reader, fileSize, miniosdk.PutObjectOptions{
		ContentType: contentType,
		UserMetadata: map[string]string{
			"Name":      fileName,
			"User-Uuid": userUUID,
		},
	})
	if err != nil {
		return fmt.Errorf("upload file %s to bucket %s: %w", fileID, bucketName, err)
	}

	return nil
}

func (c *Client) DeleteFile(ctx context.Context, bucketName, fileID string) (err error) {
	reqCtx, cancel := context.WithTimeout(resolveContext(ctx), 10*time.Second)
	defer cancel()

	if err = c.minioClient.RemoveObject(reqCtx, bucketName, fileID, miniosdk.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete file %s from bucket %s: %w", fileID, bucketName, err)
	}

	return nil
}

func resolveContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func objectName(info miniosdk.ObjectInfo) string {
	if info.UserMetadata["Name"] != "" {
		return info.UserMetadata["Name"]
	}
	return info.Key
}

func objectUserUUID(info miniosdk.ObjectInfo) string {
	return info.UserMetadata["User-Uuid"]
}

func objectContentType(info miniosdk.ObjectInfo) string {
	if info.ContentType != "" {
		return info.ContentType
	}
	return "application/octet-stream"
}
