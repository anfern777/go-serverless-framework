package repository

import (
	"context"

	"github.com/anfern777/go-serverless-framework-api/internal/models"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type BaseRepository[T any] interface {
	Save(ctx context.Context, item T) error
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string) (*T, error)
	UpdateProperty(id string, property string, propertyType string, atValue types.AttributeValue) error
	GetBetweenDates(ctx context.Context, start, end string, prefix string) ([]T, error)
	Index() *string
}

type ApplicationsRepository interface {
	BaseRepository[models.Application]
	GetApplicationIDByCognitoID(cognitoID string) (string, error)
	GetApplicationByEmail(email string) (any, error)
}

type PostsRepository interface {
	BaseRepository[models.Post]
	GetByLanguage(language string) ([]models.Post, error)
}

type ChildRepository[T any] interface {
	Save(ctx context.Context, item T) error
	Delete(parentPK string, SK string) error
	UpdateProperty(parentPK string, SK, property string, propertyType string, atValue types.AttributeValue) error
	BatchDelete(parentPK string) ([]T, error)
	BatchSave(items []T) error
	GetAllByParentPK(parentPK string) ([]T, error)
	GetByPrimaryKey(parentPK string, SK string) (*T, error)
	Index() string
}
