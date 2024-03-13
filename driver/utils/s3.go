package utils

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"k8s.io/klog/v2"
	"net/url"
	"sync/atomic"
)

/**
 * @author: HuaiAn xu
 * @date: 2024-03-12 21:14:19
 * @file: s3.go
 * @description: s3客户端
 */

// S3Client is a client
type S3Client struct {
	Config *Config
	Minio  *minio.Client
	Ctx    context.Context
}

// Metadata holds the metadata of a volume
type Metadata struct {
	BucketName    string   `json:"Name"`
	Prefix        string   `json:"Prefix"`
	Mounter       string   `json:"Mounter"`
	MountOptions  []string `json:"MountOptions"`
	CapacityBytes int64    `json:"CapacityBytes"`
}

// Config holds values to configure the driver
type Config struct {
	AccessKeyID     string
	SecretAccessKey string
	Region          string
	Endpoint        string
	Mounter         string
}

// NewS3Client creates a new S3Client
func NewS3Client(cfg *Config) (*S3Client, error) {
	var client = &S3Client{}

	client.Config = cfg
	u, err := url.Parse(client.Config.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("create Client: failed to parse endpoint: %w", err)
	}

	ssl := u.Scheme == "https"
	endpoint := u.Host
	if u.Port() == "" {
		endpoint = u.Host + ":" + u.Port()
	}

	// Create a new S3 client
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(client.Config.AccessKeyID, client.Config.SecretAccessKey, ""),
		Secure: ssl,
		Region: client.Config.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("reate Client: failed to create client: %w", err)
	}

	client.Minio = minioClient
	client.Ctx = context.Background()
	return client, nil
}

// NewClientFromSecrets creates a new S3Client from secrets
func NewClientFromSecrets(secrets map[string]string) (*S3Client, error) {
	cfg := &Config{
		AccessKeyID:     secrets["accessKeyID"],
		SecretAccessKey: secrets["secretAccessKey"],
		Region:          secrets["region"],
		Endpoint:        secrets["endpoint"],
		Mounter:         "",
	}
	return NewS3Client(cfg)
}

// CreateBucket creates a new bucket
func (c *S3Client) CreateBucket(bucketName string) error {
	err := c.Minio.MakeBucket(c.Ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		return fmt.Errorf("CreateBucket: failed to create bucket %s: %w", bucketName, err)
	}
	return nil
}

// CreatePrefix creates a new prefix
func (c *S3Client) CreatePrefix(bucketName, prefix string) error {
	_, err := c.Minio.PutObject(c.Ctx, bucketName, prefix, nil, 0, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("CreatePrefix: failed to create prefix %s: %w", prefix, err)
	}
	return nil
}

// DeleteBucket deletes a bucket
func (c *S3Client) DeleteBucket(bucketName string) error {
	var err error
	if err := c.deleteObjects(bucketName, ""); err == nil {
		return c.Minio.RemoveBucket(c.Ctx, bucketName)
	}

	klog.Warningf("DeleteBucket: failed to delete bucket %s, will try deleteObjectsOneByOne", bucketName)

	if err = c.deleteObjectsOneByOne(bucketName, ""); err == nil {
		return c.Minio.RemoveBucket(c.Ctx, bucketName)
	}

	return err
}

// DeletePrefix deletes a prefix
func (c *S3Client) DeletePrefix(bucketName, prefix string) error {
	var err error
	if err = c.Minio.RemoveObject(c.Ctx, bucketName, prefix, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("DeletePrefix: failed to delete prefix %s: %w", prefix, err)
	}

	klog.Warningf("DeletePrefix: prefix %s is deleted", prefix)

	if err = c.deleteObjectsOneByOne(bucketName, ""); err == nil {
		return c.Minio.RemoveBucket(c.Ctx, bucketName)
	}
	return err
}

// IsBucketExist checks if a bucket exists
func (c *S3Client) IsBucketExist(bucketName string) (bool, error) {
	_, err := c.Minio.BucketExists(c.Ctx, bucketName)
	if err != nil {
		return false, fmt.Errorf("IsBucketExist: failed to check if bucket %s exists: %w", bucketName, err)
	}
	return true, nil
}

// deleteObjects deletes all objects in a bucket
func (c *S3Client) deleteObjects(bucketName, prefix string) error {
	objectsCh := make(chan minio.ObjectInfo)
	var listErr error

	go func() {
		defer close(objectsCh)

		for object := range c.Minio.ListObjects(
			c.Ctx,
			bucketName,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			objectsCh <- object
		}
	}()

	if listErr != nil {
		klog.Error("Error listing objects", listErr)
		return listErr
	}

	opts := minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}
	errorCh := c.Minio.RemoveObjects(c.Ctx, bucketName, objectsCh, opts)
	haveErrWhenRemoveObjects := false
	for e := range errorCh {
		klog.Errorf("Failed to remove object %s, error: %s", e.ObjectName, e.Err)
		haveErrWhenRemoveObjects = true
	}
	if haveErrWhenRemoveObjects {
		return fmt.Errorf("failed to remove all objects of bucket %s", bucketName)
	}

	return nil
}

// will delete files one by one without file lock
func (c *S3Client) deleteObjectsOneByOne(bucketName, prefix string) error {
	parallelism := 16
	objectsCh := make(chan minio.ObjectInfo, parallelism)
	guardCh := make(chan int, parallelism)
	var listErr error
	var totalObjects int64 = 0
	var removeErrors int64 = 0

	go func() {
		defer close(objectsCh)

		for object := range c.Minio.ListObjects(c.Ctx, bucketName,
			minio.ListObjectsOptions{Prefix: prefix, Recursive: true}) {
			if object.Err != nil {
				listErr = object.Err
				return
			}
			atomic.AddInt64(&totalObjects, 1)
			objectsCh <- object
		}
	}()

	if listErr != nil {
		klog.Error("Error listing objects", listErr)
		return listErr
	}

	for object := range objectsCh {
		guardCh <- 1
		go func(obj minio.ObjectInfo) {
			err := c.Minio.RemoveObject(c.Ctx, bucketName, obj.Key,
				minio.RemoveObjectOptions{VersionID: obj.VersionID})
			if err != nil {
				klog.Errorf("Failed to remove object %s, error: %s", obj.Key, err)
				atomic.AddInt64(&removeErrors, 1)
			}
			<-guardCh
		}(object)
	}
	for i := 0; i < parallelism; i++ {
		guardCh <- 1
	}
	for i := 0; i < parallelism; i++ {
		<-guardCh
	}

	if removeErrors > 0 {
		return fmt.Errorf("failed to remove %v objects out of total %v of path %s", removeErrors, totalObjects, bucketName)
	}

	return nil
}
