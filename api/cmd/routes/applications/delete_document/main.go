package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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
	fsService                 services.FileStorageService
	databaseService           services.DatabaseClientAPI
	documentRepositoryService *repository.ChildRepositoryImpl[models.Document]

	logger     = implementations.NewSlogLogger()
	bucketName = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
)

func init() {
	var err error
	fsService, err = implementations.NewS3FileStorageService(bucketName, implementations.WithS3PresignClient())
	if err != nil {
		log.Fatalf("failed to initialize s3 service - execution stopped")
	}
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, documentType, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			fmt.Sprintf("Invalid request path parameter: %s", err),
			logger,
		)
	}

	document, err := documentRepositoryService.GetByPrimaryKey(applicationId, models.CreateDocumentSK(*documentType))
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to get document: %v", err), logger)
	}

	err = documentRepositoryService.Delete(document.PK, document.SK)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete document: %v", err), logger)
	}

	err = fsService.Delete(*document.Name)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to delete document from s3: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request *events.APIGatewayV2HTTPRequest) (string, *models.DocumentType, error) {
	id, exists := request.PathParameters["id"]
	if !exists || id == "" {
		return "", nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}
	id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")

	documentType, exists := request.PathParameters["document-type"]
	if !exists || id == "" {
		return "", nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}

	return id, common.PtrTo(models.DocumentType(documentType)), nil
}
