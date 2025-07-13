package main

import (
	"context"
	_ "embed"
	"encoding/json"
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

type DocumentData struct {
	DocumentType models.DocumentType `json:"documentType"`
	Note         string              `json:"note"`
}

type RequestBody struct {
	Documents []DocumentData `json:"documents"`
}

var (
	err                          error
	databaseService              services.DatabaseClientAPI
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]
	applicationRepositoryService *repository.ApplicationsRepositoryImpl

	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	index     = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
	// applicationEmailAddress = os.Getenv("APPLICATION_EMAIL")
	logger    = implementations.NewSlogLogger()
	documents []models.Document

	// // go:embed email-template/message-user.html
	// messageUserTemplate string
)

func init() {
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName, repository.WithGSIIndex[models.Application](index))
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, documentsData, err := processRequest(&request, ctx.Value(middleware.USER_GROUP_KEY).(middleware.UserGroup))
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest,
			fmt.Sprintf("Invalid request: %v", err),
			logger,
		)
	}

	application, err := applicationRepositoryService.GetByID(context.TODO(), applicationId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest,
			fmt.Sprintf("Invalid ID provided:%v", err),
			logger,
		)
	}

	for _, docData := range documentsData {
		doc := models.Document{
			Notes: docData.Note,
		}
		doc.GenerateKeys(application.PK, docData.DocumentType)
		if err = doc.GenerateAttributesForDocumentRequest(application.PK, doc.Notes); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest,
				fmt.Sprintf("Failed to generate attributes: %v", err),
				logger,
			)
		}
		if err := doc.Validate(); err != nil {
			return common.RequestErrorResponse(http.StatusBadRequest,
				fmt.Sprintf("Validation error: %v", err),
				logger,
			)
		}
		documents = append(documents, doc)
	}

	if err = documentRepositoryService.BatchSave(documents); err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			fmt.Sprintf("Failed to save documents in database: %v", err),
			logger,
		)
	}

	// notify user
	// if emailService, err := implementations.NewSesEmail(); err != nil {
	// 	logger.Log(fmt.Sprintf("Failed to get emailService: %v", err), services.SEVERITY_ERROR)
	// } else {
	// 	if err = emailService.Email(messageUserTemplate,
	// 		"go-serverless-framework - New Message Received",
	// 		applicationEmailAddress,
	// 		application.Email,
	// 		map[string]string{
	// 			"CurrentYear": fmt.Sprint(time.Now().Year()),
	// 		}); err != nil {
	// 		logger.Log(fmt.Sprintf("Failed to send email to user: %v", err), services.SEVERITY_ERROR)
	// 	}
	// }

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request *events.APIGatewayV2HTTPRequest, group middleware.UserGroup) (string, []DocumentData, error) {
	var requestBody RequestBody

	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse request body: %v", err)
	}

	switch group {
	case middleware.ADMIN:
		id, ok := request.PathParameters["id"]
		if !ok {
			return "", nil, fmt.Errorf("missing parameter 'id'")
		}
		id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")
		return id, requestBody.Documents, nil
	case middleware.USER:
		cognitoID, ok := request.RequestContext.Authorizer.JWT.Claims["sub"]
		if !ok {
			return "", nil, fmt.Errorf("missing sub in JWT claims")
		}
		id, err := applicationRepositoryService.GetApplicationIDByCognitoID(cognitoID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get Application ID by CognitoID")
		}
		return id, requestBody.Documents, nil
	default:
		return "", nil, errors.New("process request failed: Invalid user group")
	}
}
