package implementations

import (
	"context"
	"errors"
	"testing"

	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type MockS3ClientAPI struct {
	deleteObjectFuncCalled  bool
	deleteObjectsFuncCalled bool
}

func (msa *MockS3ClientAPI) DeleteObject(ctx context.Context, params *s3.DeleteObjectInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	msa.deleteObjectFuncCalled = true
	return nil, nil
}
func (msa *MockS3ClientAPI) DeleteObjects(ctx context.Context, params *s3.DeleteObjectsInput, optFns ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error) {
	msa.deleteObjectFuncCalled = true
	return nil, nil
}

type MockS3PresignClientAPI struct {
	PresignGetObjectCalls []struct {
		Bucket string
		Key    string
		OptFns []func(*s3.PresignOptions)
	}
	PresignGetObjectFunc func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
	PresignPutObjectFunc func(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

func (m *MockS3PresignClientAPI) PresignGetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	m.PresignGetObjectCalls = append(m.PresignGetObjectCalls, struct {
		Bucket string
		Key    string
		OptFns []func(*s3.PresignOptions)
	}{
		Bucket: *params.Bucket,
		Key:    *params.Key,
		OptFns: optFns,
	})
	return m.PresignGetObjectFunc(ctx, params, optFns...)
}

func (m *MockS3PresignClientAPI) PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	m.PresignGetObjectCalls = append(m.PresignGetObjectCalls, struct {
		Bucket string
		Key    string
		OptFns []func(*s3.PresignOptions)
	}{
		Bucket: *params.Bucket,
		Key:    *params.Key,
		OptFns: optFns,
	})
	return m.PresignPutObjectFunc(ctx, params, optFns...)
}

type PresignGetObjectError struct{}

func (pgoe *PresignGetObjectError) Error() string {
	return "error from func PresignedGetObject"
}

func TestGrantReadAccess(t *testing.T) {
	tests := []struct {
		name                         string
		key                          string
		bucket                       string
		presignFunc                  func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
		expectError                  bool
		expectedPresignFuncCallCount int
	}{
		{
			name:   "happy path",
			key:    "example-key",
			bucket: "test-bucket",
			presignFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
				return &v4.PresignedHTTPRequest{
					URL:          "test-url.com",
					Method:       "PUT",
					SignedHeader: nil,
				}, nil
			},
			expectError:                  false,
			expectedPresignFuncCallCount: 1,
		},
		{
			name:   "PresignGetObject returns error",
			key:    "example-key",
			bucket: "test-bucket",
			presignFunc: func(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
				return nil, errors.New("test this error!")
			},
			expectError:                  true,
			expectedPresignFuncCallCount: 1,
		},
	}

	for _, tt := range tests {
		mockS3ClientAPI := &MockS3ClientAPI{}
		mockS3PresignClientAPI := &MockS3PresignClientAPI{
			PresignGetObjectFunc: tt.presignFunc,
		}
		s3fs := &S3FileStorageService{
			bucket:        tt.bucket,
			client:        mockS3ClientAPI,
			presignClient: mockS3PresignClientAPI,
		}

		t.Run(tt.name, func(t *testing.T) {
			_, err := s3fs.GrantReadAccess(tt.key)
			if err != nil && !tt.expectError {
				t.Errorf("Expected no error, but got %v", err)
			}
			if err == nil && tt.expectError {
				t.Errorf("Expected error %v, but got no error", err)
			}
			presignedGetObjectCallsCount := len(mockS3PresignClientAPI.PresignGetObjectCalls)
			if tt.expectedPresignFuncCallCount != presignedGetObjectCallsCount {
				t.Errorf("expected PresignGetObject to be called %d times, but got called %d times", tt.expectedPresignFuncCallCount, presignedGetObjectCallsCount)
			}
		})
	}
}
