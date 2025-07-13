package implementations

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func NewDynamodbDatabaseService(table string) (*dynamodb.Client, error) {
	dynamodbClient, err := GetDynamoDbClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get dynamodb client")
	}

	return dynamodbClient, nil
}

func GetDynamoDbClient() (*dynamodb.Client, error) {
	awsConfig, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load dynamodb config: %v", err)
	}

	return dynamodb.NewFromConfig(awsConfig), nil
}
