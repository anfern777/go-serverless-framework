package repository

import (
	"context"
	"fmt"

	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

// implements BaseRepository
type BaseRepositoryImpl[T any] struct {
	databaseAPI services.DatabaseClientAPI
	table       string
	index       *string
}

type ParentRepositoryOption[T any] func(pir *BaseRepositoryImpl[T])

func WithGSIIndex[T any](index string) ParentRepositoryOption[T] {
	return func(pir *BaseRepositoryImpl[T]) {
		pir.index = &index
	}
}

func NewBaseRepositoryImpl[T any](databaseApi services.DatabaseClientAPI, table string, opts ...ParentRepositoryOption[T]) *BaseRepositoryImpl[T] {
	pir := BaseRepositoryImpl[T]{
		table:       table,
		databaseAPI: databaseApi,
	}

	for _, opt := range opts {
		opt(&pir)
	}
	return &pir
}

func (pir *BaseRepositoryImpl[T]) Index() *string {
	if pir.index == nil {
		return nil
	}
	return pir.index
}

func (pir *BaseRepositoryImpl[T]) Save(ctx context.Context, item T) error {
	avKeys, err := attributevalue.MarshalMap(item)
	if err != nil {
		return fmt.Errorf("failed to map go struct to dynamodb attribute value: %w", err)
	}

	_, err = pir.databaseAPI.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: &pir.table,
		Item:      avKeys,
	})

	if err != nil {
		return fmt.Errorf("failed to save item: %w", err)
	}
	return nil
}

func (pir *BaseRepositoryImpl[T]) Delete(ctx context.Context, id string) error {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": id,
			"SK": id,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to avMarshal: %w", err)
	}
	_, err = pir.databaseAPI.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName:           &pir.table,
		Key:                 avKeys,
		ReturnValues:        types.ReturnValueAllOld,
		ConditionExpression: aws.String("attribute_exists(PK)"),
	})

	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
}

func (pir *BaseRepositoryImpl[T]) GetByID(ctx context.Context, id string) (*T, error) {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": id,
			"SK": id,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to avMarshal: %v", err)
	}

	result, err := pir.databaseAPI.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: &pir.table,
		Key:       avKeys,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %v", err)
	}
	if result.Item == nil {
		return nil, fmt.Errorf("item not found")
	}

	var item T
	err = attributevalue.UnmarshalMap(result.Item, &item)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal map result: %v", err)
	}

	return &item, nil
}

func (pir *BaseRepositoryImpl[T]) GetBetweenDates(ctx context.Context, start, end string, prefix string) ([]T, error) {
	if pir.index == nil {
		return nil, fmt.Errorf("database index is not set")
	}
	queryOutput, err := pir.databaseAPI.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:              &pir.table,
		IndexName:              pir.index,
		KeyConditionExpression: aws.String("GSI_PK = :gsi_pk and CreatedAt between :s AND :e"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":gsi_pk": &types.AttributeValueMemberS{Value: prefix},
			":s":      &types.AttributeValueMemberS{Value: start},
			":e":      &types.AttributeValueMemberS{Value: end},
		},
		ScanIndexForward: aws.Bool(false),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query table: %w", err)
	}

	var items []T
	err = attributevalue.UnmarshalListOfMaps(queryOutput.Items, &items)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal query output: %w", err)
	}
	return items, nil
}

func (pir *BaseRepositoryImpl[T]) UpdateProperty(id string, property string, propertyType string, atValue types.AttributeValue) error {
	avKeys, err := attributevalue.MarshalMap(
		map[string]string{
			"PK": id,
			"SK": id,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to avMarshal: %w", err)
	}

	_, err = pir.databaseAPI.UpdateItem(context.TODO(), &dynamodb.UpdateItemInput{
		TableName:        &pir.table,
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
