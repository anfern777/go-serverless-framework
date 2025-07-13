package repository

import (
	"context"
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/models"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestNewApplicationsRepositoryImpl(t *testing.T) {
	mockDb := MockDatabaseAPI{
		GetItemFunc: func(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
			return &dynamodb.GetItemOutput{
				Item: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "test-pk"},
					"SK": &types.AttributeValueMemberS{Value: "test-sk"},
				},
			}, nil
		},
	}
	index := "test-index"
	table := "test-table"
	gtRepo := NewApplicationsRepositoryImpl(&mockDb, table, WithGSIIndex[models.Application](index))
	if *gtRepo.Index() != "test-index" {
		t.Errorf("Expected index to be '%s', but got %s", index, *gtRepo.Index())
	}
	if gtRepo.table != "test-table" {
		t.Errorf("Expected table to be '%s', but got %s", table, gtRepo.table)
	}
	item, err := gtRepo.GetByID(context.TODO(), table)
	if err != nil {
		t.Errorf("Expected no error, but got %v", err)
	}
	if item == nil {
		t.Errorf("Expected item to not be nil")
	}
}
