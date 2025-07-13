package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/implementations"
	"github.com/anfern777/go-serverless-framework-api/internal/middleware"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	dyndb "github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	err             error
	databaseService services.DatabaseClientAPI

	logger    services.Logger
	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	postId, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			"Invalid request path parameter",
			logger,
		)
	}

	// Unmarshal request body
	var updatedPost models.Post
	if err := json.Unmarshal([]byte(request.Body), &updatedPost); err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			fmt.Sprintf("Invalid request body: %v", err),
			logger,
		)
	}

	if updatedPost.Title == "" || updatedPost.Content == "" {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			"Invalid request body - empty fields",
			logger,
		)
	}
	avKeys, err := attributevalue.MarshalMap(map[string]string{
		"PK": *postId,
		"SK": *postId,
	})
	if err != nil {
		return nil, err
	}

	updateOutput, err := databaseService.UpdateItem(ctx, &dyndb.UpdateItemInput{
		Key:              avKeys,
		TableName:        aws.String(tableName),
		UpdateExpression: aws.String("SET Title = :t, Content = :c"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":t": &types.AttributeValueMemberS{Value: updatedPost.Title},
			":c": &types.AttributeValueMemberS{Value: updatedPost.Content},
		},
		ReturnValues: types.ReturnValueAllNew,
	})
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusInternalServerError,
			fmt.Sprintf("Update failed: %v", err),
			logger,
		)
	}
	if updateOutput == nil || len(updateOutput.Attributes) == 0 {
		return common.RequestErrorResponse(
			http.StatusNotFound,
			"Item not found or update did not succeed",
			logger,
		)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request *events.APIGatewayV2HTTPRequest) (*string, error) {
	encodedPostId, exists := request.PathParameters["id"]
	if !exists || encodedPostId == "" {
		return nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}

	postId, err := url.QueryUnescape(encodedPostId)
	if err != nil {
		return nil, fmt.Errorf("failed to decode id from request: %w", err)
	}

	return &postId, nil
}
