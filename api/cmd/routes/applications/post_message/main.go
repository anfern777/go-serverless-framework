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

var (
	err                          error
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	messageRepositoryService     *repository.ChildRepositoryImpl[models.Message]

	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	indexName = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
	logger    = implementations.NewSlogLogger()

	// applicationEmail = os.Getenv("APPLICATION_EMAIL")
	// officeEmail      = os.Getenv("OFFICE_EMAIL")

	// // go:embed email-template/message-admin.html
	// messageAdminTemplate string

	// // go:embed email-template/message-user.html
	// messageUserTemplate string
)

func init() {
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName, repository.WithGSIIndex[models.Application](indexName))
	messageRepositoryService = repository.NewChildRepositoryImpl[models.Message](databaseService, tableName)
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, message, err := processRequest(&req, ctx.Value(middleware.USER_GROUP_KEY).(middleware.UserGroup))
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid request - %v", err), logger)
	}

	message.GenerateKeys(applicationId)
	err = message.Validate()
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed validation - %v", err), logger)
	}

	// save application in database
	err = messageRepositoryService.Save(context.TODO(), *message)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to save application in database: %v", err), logger)
	}

	// TODO: send email to admin and applicant
	// if emailService, err := implementations.NewSesEmail(); err != nil {
	// 	logger.Log(fmt.Sprintf("Failed to get emailService: %v", err), services.SEVERITY_ERROR)
	// } else {
	// 	if err = emailService.Email(messageAdminTemplate,
	// 		"[go-serverless-framework Website] - New Message!",
	// 		applicationEmail,
	// 		officeEmail,
	// 		map[string]string{
	// 			"Email":       email,
	// 			"CurrentYear": fmt.Sprint(time.Now().Year()),
	// 		}); err != nil {
	// 		logger.Log(fmt.Sprintf("Failed to send email to admin: %v", err), services.SEVERITY_ERROR)
	// 	}
	// 	if err = emailService.Email(messageUserTemplate,
	// 		"go-serverless-framework - New Message Received",
	// 		applicationEmail,
	// 		email,
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
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN, middleware.USER))
}

func processRequest(request *events.APIGatewayV2HTTPRequest, group middleware.UserGroup) (string, *models.Message, error) {
	var message models.Message
	err := json.Unmarshal([]byte(request.Body), &message)
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse message from request")
	}

	switch group {
	case middleware.ADMIN:
		id, ok := request.PathParameters["id"]
		if !ok {
			return "", nil, fmt.Errorf("missing parameter 'id'")
		}
		id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")
		message.GenerateAttributes(models.MessageAuthor(models.MESSAGE_AUTHOR_ADMIN))
		return id, &message, nil
	case middleware.USER:
		cognitoID, ok := request.RequestContext.Authorizer.JWT.Claims["sub"]
		if !ok || cognitoID == "" {
			return "", nil, fmt.Errorf("missing sub in JWT claims")
		}
		id, err := applicationRepositoryService.GetApplicationIDByCognitoID(cognitoID)
		if err != nil {
			return "", nil, fmt.Errorf("failed to get Application ID by CognitoID=%s", cognitoID)
		}
		message.GenerateAttributes(models.MessageAuthor(models.MESSAGE_AUTHOR_USER))
		return id, &message, nil
	default:
		return "", nil, errors.New("process request failed: Invalid user group")
	}
}
