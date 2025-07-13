package repository

import (
	"context"
	"fmt"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

// implements BaseRepository
// implements PostsRepository
type PostsRepositoryImpl struct {
	ChildRepositoryImpl[models.Post]
	baseValues *ChildRepositoryImpl[models.Post]
}

func NewPostsRepositoryImpl(databaseApi services.DatabaseClientAPI, table string, opts ...ChildRepositoryImplOption[models.Post]) *PostsRepositoryImpl {
	base := NewChildRepositoryImpl(databaseApi, table, opts...)
	baseProps := &ChildRepositoryImpl[models.Post]{
		table:       table,
		databaseAPI: databaseApi,
		index:       common.PtrTo(base.Index()),
	}
	return &PostsRepositoryImpl{
		baseValues: baseProps,
	}
}

func (p *PostsRepositoryImpl) GetByLanguage(language string) ([]models.Post, error) {
	result, err := p.databaseAPI.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &p.table,
		IndexName:              p.index,
		KeyConditionExpression: aws.String("GSI_PK = :gsi_pk"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			// Use the validated 'language' (which is common.LanguageCode type) for the query
			":gsi_pk": &types.AttributeValueMemberS{Value: common.PK_POST_PREFIX + string(language) + "#"},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query table %s:%w", "posts", err)
	}

	var items []models.Post
	err = attributevalue.UnmarshalListOfMaps(result.Items, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query output %s:%w", "posts", err)
	}
	return items, nil
}
