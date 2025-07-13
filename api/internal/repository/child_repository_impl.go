package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

// implements ChildRepository
type ChildRepositoryImpl[T any] struct {
	databaseAPI services.DatabaseClientAPI
	table       string
	index       *string
}

type ChildRepositoryImplOption[T any] func(ci *ChildRepositoryImpl[T])

func WithBaseAssociationsIndex[T any](index string) ChildRepositoryImplOption[T] {
	return func(ci *ChildRepositoryImpl[T]) {
		ci.index = &index
	}
}

func NewChildRepositoryImpl[T any](databaseApi services.DatabaseClientAPI, table string, opts ...ChildRepositoryImplOption[T]) *ChildRepositoryImpl[T] {
	d := &ChildRepositoryImpl[T]{
		table:       table,
		databaseAPI: databaseApi,
	}

	for _, opt := range opts {
		opt(d)
	}

	return &ChildRepositoryImpl[T]{
		table:       table,
		databaseAPI: databaseApi,
	}
}

func (ci *ChildRepositoryImpl[T]) Index() string {
	return *ci.index
}

func (ci *ChildRepositoryImpl[T]) Save(ctx context.Context, item T) error {
	avItem, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to marshal map item: %w", err)
	}
	parentPKValue, ok := avItem["PK"]
	if !ok {
		return errors.New("marshaled item is missing a PK")
	}

	parentSKValue := parentPKValue

	_, err = ci.databaseAPI.TransactWriteItems(context.TODO(), &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				ConditionCheck: &types.ConditionCheck{
					TableName: aws.String(ci.table),
					Key: map[string]types.AttributeValue{
						"PK": parentPKValue,
						"SK": parentSKValue,
					},
					ConditionExpression: aws.String("attribute_exists(PK)"),
				},
			},
			{
				Put: &types.Put{
					TableName:           aws.String(ci.table),
					Item:                avItem,
					ConditionExpression: aws.String("attribute_not_exists(SK)"),
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to save item in dynamoDb due to transaction failure: %w", err)
	}

	return nil
}

func (ci *ChildRepositoryImpl[T]) Delete(parentPK string, SK string) error {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": parentPK,
			"SK": SK,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to avMarshal: %w", err)
	}

	_, err = ci.databaseAPI.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: &ci.table,
		Key:       avKeys,
	})
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

func (ci *ChildRepositoryImpl[T]) GetAllByParentPK(parentPK string) ([]T, error) {
	var zeroT T
	var prefix string
	switch (any(zeroT)).(type) {
	case models.Document:
		prefix = common.DOCUMENT_LETTER_PREFIX
	case models.Message:
		prefix = common.MESSAGE_LETTER_PREFIX
	default:
		return nil, fmt.Errorf("invalid type under type switch")
	}

	queryInput := &dynamodb.QueryInput{
		TableName:              &ci.table,
		KeyConditionExpression: aws.String("PK = :pk and begins_with ( SK, :prefix )"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":pk":     &types.AttributeValueMemberS{Value: parentPK},
			":prefix": &types.AttributeValueMemberS{Value: prefix},
		},
	}

	response, err := ci.databaseAPI.Query(context.TODO(), queryInput)
	if err != nil {
		return nil, fmt.Errorf("failed to query item items: %w", err)
	}

	var result []T
	attributevalue.UnmarshalListOfMaps(response.Items, &result)

	return result, nil
}

func (ci *ChildRepositoryImpl[T]) BatchDelete(parentPK string) ([]T, error) {
	items, err := ci.GetAllByParentPK(parentPK)
	if err != nil {
		return nil, fmt.Errorf("failed to query item items: %w", err)
	}

	if len(items) == 0 {
		return nil, nil
	}

	writeRequests := make([]types.WriteRequest, 0, len(items))
	for _, item := range items {
		sk, err := common.GetValueFromGenericStruct(item, "SK")
		if err != nil {
			return nil, fmt.Errorf("failed to get SK from item: %w", err)
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			DeleteRequest: &types.DeleteRequest{
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: parentPK},
					"SK": &types.AttributeValueMemberS{Value: sk.(string)},
				},
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			ci.table: writeRequests,
		},
	}

	_, err = ci.databaseAPI.BatchWriteItem(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to batch delete items: %w", err)
	}

	return items, nil
}

func (ci *ChildRepositoryImpl[T]) BatchSave(items []T) error {
	var avItem map[string]types.AttributeValue
	var err error

	writeRequests := make([]types.WriteRequest, 0, len(items))
	for _, doc := range items {
		avItem, err = attributevalue.MarshalMap(doc)
		if err != nil {
			return fmt.Errorf("failed to MarshalMap item: %w", err)
		}
		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: avItem,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			ci.table: writeRequests,
		},
	}

	_, err = ci.databaseAPI.BatchWriteItem(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to batch save items in dynamoDb: %w", err)
	}

	return nil
}

func (ci *ChildRepositoryImpl[T]) UpdateProperty(parentPK string, SK, property string, propertyType string, atValue types.AttributeValue) error {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": parentPK,
			"SK": parentPK,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to avMarshal: %w", err)
	}

	_, err = ci.databaseAPI.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:        &ci.table,
		Key:              avKeys,
		UpdateExpression: aws.String(fmt.Sprintf("SET %s = :%s", property, propertyType)),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			fmt.Sprintf(":%s", propertyType): atValue,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}
	return nil
}

func (ci *ChildRepositoryImpl[T]) GetByPrimaryKey(parentPK string, SK string) (*T, error) {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": parentPK,
			"SK": SK,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to marshalmap model key: %w", err)
	}

	result, err := ci.databaseAPI.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &ci.table,
		Key:       avKeys,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve item: %w", err)
	}
	if result == nil || len(result.Item) == 0 {
		return nil, errors.New("failed to retrieve item - does not exist")
	}

	var item T
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to UnmarshalMap applicationItem: %w", err)
	}

	return &item, nil
}
