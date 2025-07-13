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
	"github.com/google/uuid"
)

// implements BaseRepository
// implements ApplicationsRepository
type ApplicationsRepositoryImpl struct {
	BaseRepositoryImpl[models.Application]
}

func NewApplicationsRepositoryImpl(databaseApi services.DatabaseClientAPI, table string, opts ...ParentRepositoryOption[models.Application]) *ApplicationsRepositoryImpl {
	base := NewBaseRepositoryImpl(databaseApi, table, opts...)
	return &ApplicationsRepositoryImpl{
		BaseRepositoryImpl: BaseRepositoryImpl[models.Application]{
			databaseAPI: databaseApi,
			table:       table,
			index:       base.Index(),
		},
	}
}

func (a *ApplicationsRepositoryImpl) GetApplicationByEmail(email string) (any, error) {
	result, err := a.databaseAPI.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &a.table,
		IndexName:              a.index,
		KeyConditionExpression: aws.String("Email = :email"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":email": &types.AttributeValueMemberS{Value: email},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query item: %w", err)
	}
	if result == nil || len(result.Items) != 1 {
		return nil, fmt.Errorf("item not found: %v", err)
	}

	var app models.Application
	err = attributevalue.UnmarshalMap(result.Items[0], &app)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal output: %v", err)
	}

	return &app, nil
}

func (a *ApplicationsRepositoryImpl) GetApplicationIDByCognitoID(cognitoID string) (string, error) {
	queryInput := &dynamodb.QueryInput{
		TableName:              &a.table,
		IndexName:              a.index,
		KeyConditionExpression: aws.String("CognitoID = :cid"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":cid": &types.AttributeValueMemberS{Value: cognitoID},
		},
	}

	response, err := a.databaseAPI.Query(context.TODO(), queryInput)
	if err != nil {
		return "", fmt.Errorf("failed to query item: %w", err)
	}

	var result models.Application
	attributevalue.UnmarshalMap(response.Items[0], &result)
	if len(response.Items) == 0 || len(response.Items) != 1 {
		return "", fmt.Errorf("empty or non-unique result for cognitoID %s", cognitoID)
	}

	return result.PK, nil
}

func (a *ApplicationsRepositoryImpl) Delete(ctx context.Context, id string) error {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": id,
			"SK": id,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to avMarshal: %w", err)
	}
	app, err := a.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item for posterior deletion: %w", err)
	}
	_, err = a.databaseAPI.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		ClientRequestToken: aws.String(uuid.New().String()),
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName:           &a.table,
					Key:                 avKeys,
					ConditionExpression: aws.String("attribute_exists(PK)"),
				},
			},
			{
				Delete: &types.Delete{
					TableName: &a.table,
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{
							Value: fmt.Sprintf("%s%s%s%s%s", app.GSI_PK, common.PK_SEPARATOR, "email", common.PK_SEPARATOR, app.Email),
						},
						"SK": &types.AttributeValueMemberS{
							Value: "EMAIL_UNIQUE_CONSTRAINT",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}
	return nil
}

func (a *ApplicationsRepositoryImpl) Save(ctx context.Context, item models.Application) error {
	avKeys, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to map go struct to dynamodb attribute value: %w", err)
	}

	// special case for Application
	_, err = a.databaseAPI.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		ClientRequestToken: aws.String(uuid.New().String()),
		TransactItems: []types.TransactWriteItem{
			{
				Put: &types.Put{
					TableName:           &a.table,
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
					Item:                avKeys,
				},
			},
			{
				Put: &types.Put{
					TableName:           &a.table,
					ConditionExpression: aws.String("attribute_not_exists(PK)"),
					Item: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{
							Value: fmt.Sprintf("%s%s%s%s%s", item.GSI_PK, common.PK_SEPARATOR, "email", common.PK_SEPARATOR, item.Email),
						},
						"SK": &types.AttributeValueMemberS{
							Value: "EMAIL_UNIQUE_CONSTRAINT",
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to TransactWriteItems: %w", err)
	}
	return nil
}
