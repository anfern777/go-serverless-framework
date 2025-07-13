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
	"github.com/anfern777/go-serverless-framework-api/internal/repository"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	err                   error
	databaseService       services.DatabaseClientAPI
	postRepositoryService *repository.BaseRepositoryImpl[models.Post]

	logger    services.Logger
	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()

	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	postRepositoryService = repository.NewBaseRepositoryImpl[models.Post](databaseService, tableName)
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

	// get post data from dynamoDb
	post, err := postRepositoryService.GetByID(context.TODO(), *postId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to get item: %v", err), logger)
	}

	jsonPost, err := json.Marshal(post)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal scan output items: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonPost),
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
