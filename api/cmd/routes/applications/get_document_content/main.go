package main

import (
	"context"
	"encoding/base64"
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
	err                       error
	fsService                 services.FileStorageService
	databaseService           services.DatabaseClientAPI
	documentRepositoryService *repository.ChildRepositoryImpl[models.Document]

	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	bucketName = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
	logger     = implementations.NewSlogLogger()
)

func init() {
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
			"Invalid request path parameter",
			logger,
		)
	}

	if *documentType != models.DocumentTypeCV && *documentType != models.DocumentTypeConsent {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid docType sent in request: %s", fmt.Sprint(*documentType)), logger)
	}

	document, err := documentRepositoryService.GetByPrimaryKey(applicationId, models.CreateDocumentSK(*documentType))
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to get document from database: %v", err), logger)
	}

	if document.UploadedAt == nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, "Failed to get document content - file not uploaded: %v", logger)
	}

	pr, err := fsService.GrantReadAccess(*document.Name)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to generate signed URL: %v", err), logger)
	}

	fileContent, err := common.GetFileContent(pr.URL)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to get file content: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": *document.ContentType,
		},
		IsBase64Encoded: true,
		Body:            base64.StdEncoding.EncodeToString(fileContent),
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
	if !exists || documentType == "" {
		return "", nil, fmt.Errorf("failed to get document type from path parameter: %w", err)
	}
	return id, common.PtrTo(models.DocumentType(documentType)), nil
}
