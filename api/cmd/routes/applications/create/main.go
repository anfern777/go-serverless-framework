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
	Application   models.Application `json:"application"`
	DocumentsData []DocumentData     `json:"documents"`
}

type ResponseBody struct {
	Application       models.Application         `json:"application"`
	PresignedRequests []*v4.PresignedHTTPRequest `json:"presignedRequests"`
}

var (
	err                          error
	fsService                    services.FileStorageService
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]
	queuingService               services.MessageQueuing[services.EmailerInputParams]

	tableName        = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	logger           services.Logger
	officeEmail      = os.Getenv("OFFICE_EMAIL")
	applicationEmail = os.Getenv("APPLICATION_EMAIL")
	bucketName       = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
	emailQueueURL    = os.Getenv(common.LAMBDA_ENV_SQS_QUEUE_URL)

	//go:embed email_templates/admin_email.html
	adminEmailTemplate string

	//go:embed email_templates/user_email.html
	userEmailTemplate string
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
	queuingService, err = implementations.NewSQSService[services.EmailerInputParams](emailQueueURL)
	if err != nil {
		logger.Log(
			"Failed to initialize queuing service",
			slog.LevelWarn,
			"Error", err,
		)
	}
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var application models.Application
	var documents []models.Document
	var requestBody RequestBody
	var responseBody *ResponseBody

	err := json.Unmarshal([]byte(req.Body), &requestBody)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to parse request body: %v", err), logger)
	}

	application = requestBody.Application
	documentsData := requestBody.DocumentsData

	application.GenerateKeys()
	application.GenerateAttributes()

	for _, docData := range documentsData {
		doc := models.Document{
			SizeBytes: &docData.SizeBytes,
			Name:      &docData.Name,
		}
		doc.GenerateKeys(application.PK, docData.DocumentType)
		if err = doc.GenerateAttributesForDocumentRequest(application.PK, ""); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate attributes for document request: %v", err), logger)
		}
		if err = doc.GenerateAttributesForDocumentUpload(*doc.Name, *doc.SizeBytes); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate attributes for document upload: %v", err), logger)
		}
		if err = doc.Validate(); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed validation: %v", err), logger)
		}
		documents = append(documents, doc)
	}

	err = application.Validate(documents)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed validation - %v", err), logger)
	}

	err = applicationRepositoryService.Save(context.TODO(), application)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to save application in database: %v", err), logger)
	}

	err = documentRepositoryService.BatchSave(documents)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to save documents in database: %v", err), logger)
	}

	presignedRequests := make([]*v4.PresignedHTTPRequest, 0, len(documents))
	for _, doc := range documents {
		pr, err := fsService.GrantWriteAccess(*doc.Name)
		if err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to generate pre-signed request: %v", err), logger)
		}
		presignedRequests = append(presignedRequests, pr)
	}

	responseBody = &ResponseBody{
		Application:       application,
		PresignedRequests: presignedRequests,
	}

	jsonResponse, err := json.Marshal(responseBody)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal response: %v", err), logger)
	}

	// send email to admin and applicant
	name, email, message := application.GetEmailData()

	if queuingService != nil {
		emailInputParams := &services.EmailerInputParams{
			Template:    adminEmailTemplate,
			Subject:     "Admin: Application Received",
			Source:      applicationEmail,
			Destination: officeEmail,
			Data: map[string]string{
				"Name":    name,
				"Email":   email,
				"Message": message,
			}}

		if err = queuingService.SendMessage(emailInputParams); err != nil {
			logger.Log(
				"Failed to send email notification to admin",
				slog.LevelError,
				"Error", err,
			)
		}
		emailInputParams = &services.EmailerInputParams{
			Template:    userEmailTemplate,
			Subject:     "go-serverless-framework: Application Received",
			Source:      applicationEmail,
			Destination: email,
			Data: map[string]string{
				"Name":    name,
				"Message": message,
			}}
		if err = queuingService.SendMessage(emailInputParams); err != nil {
			logger.Log(
				"Failed to send email notification to user",
				slog.LevelError,
				"Error", err,
			)
		}
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusCreated,
		Body:       string(jsonResponse),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
