package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"log/slog"
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

type RequestBody struct {
	Name         string `json:"name"`
	DocumentType string `json:"documentType"`
	SizeBytes    int64  `json:"sizeBytes"`
}

var (
	err                          error
	fsService                    services.FileStorageService
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]

	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	indexName  = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
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
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName, repository.WithGSIIndex[models.Application](indexName))
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, requestBody, err := processRequest(&request, ctx.Value(middleware.USER_GROUP_KEY).(middleware.UserGroup))
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err), logger)
	}

	document, err := documentRepositoryService.GetByPrimaryKey(applicationId, models.CreateDocumentSK(models.DocumentType(requestBody.DocumentType)))
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to get document: %v", err), logger)
	}

	if err = document.GenerateAttributesForDocumentUpload(requestBody.Name, requestBody.SizeBytes); err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate attributes for upload: %v", err), logger)
	}

	if err = documentRepositoryService.Save(context.TODO(), *document); err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to update document: %v", err), logger)
	}

	// add presigned s3 url to response
	presignedRequest, err := fsService.GrantWriteAccess(*document.Name)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate pre-signed request: %v", err), logger)
	}

	jsonResponse, err := json.Marshal(presignedRequest)
	if err != nil {
		logger.Log(
			fmt.Sprintf("Failed to Marshal response"),
			slog.LevelError,
			"Error", err,
		)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusCreated,
		Body:       string(jsonResponse),
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN, middleware.USER))
}

func processRequest(request *events.APIGatewayV2HTTPRequest, group middleware.UserGroup) (string, *RequestBody, error) {
	var requestBody RequestBody

	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	switch group {
	case middleware.ADMIN:
		id, ok := request.PathParameters["id"]
		if !ok {
			return "", nil, fmt.Errorf("missing parameter 'id'")
		}
		id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")
		return id, &requestBody, nil
	case middleware.USER:
		cognitoID, ok := request.RequestContext.Authorizer.JWT.Claims["sub"]
		if !ok {
			return "", nil, fmt.Errorf("missing sub in JWT claims")
		}
		id, err := applicationRepositoryService.GetApplicationIDByCognitoID(cognitoID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get Application ID by CognitoID")
		}
		return id, &requestBody, nil
	default:
		return "", nil, errors.New("process request failed: Invalid user group")
	}
}
