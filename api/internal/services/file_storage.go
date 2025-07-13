package services

import v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"

type FileStorageService interface {
	GrantReadAccess(key string) (*v4.PresignedHTTPRequest, error)
	GrantWriteAccess(key string) (*v4.PresignedHTTPRequest, error)
	Save(key string) error
	Delete(key string) error
	BatchDelete(keys []string) error
}
