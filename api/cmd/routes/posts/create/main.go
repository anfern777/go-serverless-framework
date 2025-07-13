package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
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
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
)

type DocumentData struct {
	Name         string              `json:"name"`
	SizeBytes    int64               `json:"sizeBytes"`
	DocumentType models.DocumentType `json:"documentType"`
}

type RequestBody struct {
	Post          models.Post    `json:"post"`
	DocumentsData []DocumentData `json:"documents"`
}

type ResponseBody struct {
	Post              models.Post                `json:"post"`
	PresignedRequests []*v4.PresignedHTTPRequest `json:"presignedRequests"`
}

var (
	err                       error
	fsService                 services.FileStorageService
	databaseService           services.DatabaseClientAPI
	postRepositoryService     *repository.BaseRepositoryImpl[models.Post]
	documentRepositoryService *repository.ChildRepositoryImpl[models.Document]

	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	logger     = implementations.NewSlogLogger()
	bucketName = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
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
	postRepositoryService = repository.NewBaseRepositoryImpl[models.Post](databaseService, tableName)
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var post models.Post
	var documents []models.Document
	var requestBody RequestBody
	var responseBody *ResponseBody

	language, exists := req.PathParameters["language"]
	if !exists {
		return common.RequestErrorResponse(http.StatusBadRequest, "Invalid or inexistent language in request", logger)
	}

	err := json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to parse request body: %v", err), logger)
	}

	// parse and validate post and document data structures
	post = requestBody.Post
	documentsData := requestBody.DocumentsData

	post.GenerateKeys(common.LanguageCode(language))
	post.GenerateAttributes()

	for _, docData := range documentsData {
		doc := models.Document{
			SizeBytes: &docData.SizeBytes,
			Name:      &docData.Name,
		}
		doc.GenerateKeys(post.PK, docData.DocumentType)
		if err = doc.GenerateAttributesForDocumentRequest(post.PK, ""); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate attributes: %v", err), logger)
		}
		if err = doc.GenerateAttributesForDocumentUpload(*doc.Name, *doc.SizeBytes); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate attributes: %v", err), logger)
		}
		if err = doc.Validate(); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed validation: %v", err), logger)
		}
		documents = append(documents, doc)
	}
	err = post.Validate(&documents)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed validation - %v", err), logger)
	}

	err = postRepositoryService.Save(context.TODO(), post)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to save post in database: %v", err), logger)
	}

	err = documentRepositoryService.BatchSave(documents)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to save documents in database: %v", err), logger)
	}

	// add presigned s3 url to response
	presignedRequests := make([]*v4.PresignedHTTPRequest, 0, len(documents))
	for _, doc := range documents {
		pr, err := fsService.GrantWriteAccess(*doc.Name)
		if err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate pre-signed request: %v", err), logger)
		}
		presignedRequests = append(presignedRequests, pr)
	}

	responseBody = &ResponseBody{
		Post:              post,
		PresignedRequests: presignedRequests,
	}

	jsonResponse, err := json.Marshal(responseBody)
	if err != nil {
		logger.Log(
			"Failed to Marshal response",
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
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}
