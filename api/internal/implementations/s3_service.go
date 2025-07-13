package implementations

import (
	"context"
	"fmt"
	"time"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/aws-sdk-go/aws"
)

// implements services.FileStorageService
type S3FileStorageService struct {
	client S3ClientAPI
	bucket string

	presignClient S3PresignClientAPI //optional
}

type S3ClientAPI interface {
	DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
}

type S3PresignClientAPI interface {
	PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

type S3FileStorageServiceOption func(*S3FileStorageService)

func WithS3PresignClient() S3FileStorageServiceOption {
	return func(s *S3FileStorageService) {
		s.presignClient = s3.NewPresignClient(s.client.(*s3.Client))
	}
}

func NewS3FileStorageService(bucket string, opts ...S3FileStorageServiceOption) (*S3FileStorageService, error) {
	s3Client, err := getS3Client()
	if err != nil {
		return nil, fmt.Errorf("failed to get S3FileStorageService client: %w", err)
	}
	service := S3FileStorageService{
		client: s3Client,
		bucket: bucket,
	}
	for _, opt := range opts {
		opt(&service)
	}

	return &service, nil
}

func (s *S3FileStorageService) Save(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get delete object: %w", err)
	}
	return nil
}

func (s *S3FileStorageService) Delete(key string) error {
	_, err := s.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get delete object: %w", err)
	}
	return nil
}

func (s *S3FileStorageService) BatchDelete(keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	var objectIdentifiers []types.ObjectIdentifier
	for _, key := range keys {
		objKey := key
		objectIdentifiers = append(objectIdentifiers, types.ObjectIdentifier{
			Key: &objKey,
		})
	}
	_, err := s.client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: &s.bucket,
		Delete: &types.Delete{
			Objects: objectIdentifiers,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to get delete object: %w", err)
	}
	return nil
}

func (s *S3FileStorageService) GrantReadAccess(key string) (*v4.PresignedHTTPRequest, error) {
	if s.presignClient == nil {
		return nil, fmt.Errorf("client does not have presign client initialized")
	}

	pr, err := s.presignClient.PresignGetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: &s.bucket,
		Key:    aws.String(key),
	}, s3.WithPresignExpires(60*time.Minute))
	if err != nil {
		return nil, fmt.Errorf("failed to generate signed url: %w", err)
	}
	return pr, nil
}

func (s *S3FileStorageService) GrantWriteAccess(key string) (*v4.PresignedHTTPRequest, error) {
	if s.presignClient == nil {
		return nil, fmt.Errorf("client does not have presign client initialized")
	}
	pr, err := s.presignClient.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: &s.bucket,
		Key:    &key,
	}, func(o *s3.PresignOptions) {
		o.Expires = 30 * time.Second
	})
	if err != nil {
		return nil, fmt.Errorf("failed to presign put object: %w", err)
	}
	return pr, nil
}

func getS3Client() (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load s3 config: %v", err)
	}

	return s3.NewFromConfig(cfg), nil
}
