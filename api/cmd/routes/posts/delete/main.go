package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings" // Added for string manipulation

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/implementations"
	"github.com/anfern777/go-serverless-framework-api/internal/middleware"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/repository"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/google/uuid" // Added for UUID validation
)

var (
	err                       error
	fsService                 services.FileStorageService
	databaseService           services.DatabaseClientAPI
	postRepositoryService     *repository.BaseRepositoryImpl[models.Post]
	documentRepositoryService *repository.ChildRepositoryImpl[models.Document]

	logger     services.Logger
	bucketName = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()
	fsService, err = implementations.NewS3FileStorageService(bucketName, implementations.WithS3PresignClient())
	if err != nil {
		log.Fatalf("failed to initialize s3 service - execution stopped")
	}
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	postRepositoryService = repository.NewBaseRepositoryImpl[models.Post](databaseService, tableName)
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	postId, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			fmt.Sprintf("Invalid request path parameter: %s", err),
			logger,
		)
	}

	err = postRepositoryService.Delete(context.TODO(), postId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete item: %v", err), logger)
	}
	logger.Log(
		fmt.Sprintf("Post deleted successfuly from database"),
		slog.LevelError,
		"PostID", postId,
		"Error", err,
	)

	documents, err := documentRepositoryService.BatchDelete(postId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete documents: %v", err), logger)
	}

	var keys []string
	for _, p := range documents {
		keys = append(keys, *p.Name)
	}

	err = fsService.BatchDelete(keys)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete document content from s3: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request *events.APIGatewayV2HTTPRequest) (string, error) {
	encodedPostId, exists := request.PathParameters["id"]
	if !exists || encodedPostId == "" {
		return "", errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}

	postId, err := url.QueryUnescape(encodedPostId)
	if err != nil {
		return "", fmt.Errorf("failed to decode id from request: %w", err)
	}

	// Validate the format of the decoded postId
	// It should start with common.PK_POST_PREFIX and be followed by a valid UUID.
	if !strings.HasPrefix(postId, common.PK_POST_PREFIX) {
		return "", fmt.Errorf("invalid post ID format: missing prefix %s", common.PK_POST_PREFIX)
	}

	uuidPart := strings.TrimPrefix(postId, common.PK_POST_PREFIX)
	if _, err := uuid.Parse(uuidPart); err != nil {
		return "", fmt.Errorf("invalid post ID format: invalid UUID part: %w", err)
	}

	return postId, nil
}
