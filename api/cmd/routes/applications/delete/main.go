package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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
	err                          error
	fsService                    services.FileStorageService
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]
	messageRepositoryService     *repository.ChildRepositoryImpl[models.Message]

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
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName)
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
	messageRepositoryService = repository.NewChildRepositoryImpl[models.Message](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			fmt.Sprintf("Invalid request path parameter: %s", err),
			logger,
		)
	}

	err = applicationRepositoryService.Delete(context.TODO(), *applicationId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete item: %v", err), logger)
	}

	documents, err := documentRepositoryService.BatchDelete(*applicationId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete documents: %v", err), logger)
	}

	if len(documents) > 0 {
		var keys []string
		for _, d := range documents {
			keys = append(keys, *d.Name)
		}

		err = fsService.BatchDelete(keys)
		if err != nil {
			return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete document content from s3: %v", err), logger)
		}
	}

	_, err = messageRepositoryService.BatchDelete(*applicationId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete messages: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request *events.APIGatewayV2HTTPRequest) (*string, error) {
	id, exists := request.PathParameters["id"]
	if !exists || id == "" {
		return nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}
	id = common.CreatePK(common.PK_APP_PREFIX, id)

	return &id, nil
}
