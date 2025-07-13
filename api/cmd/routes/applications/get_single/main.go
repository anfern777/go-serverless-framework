package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	utils "github.com/anfern777/go-serverless-framework-api/internal/common"
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
	logger                       services.Logger
	applicationRepositoryService *repository.ApplicationsRepositoryImpl

	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	indexName = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName, repository.WithGSIIndex[models.Application](indexName))
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, err := processRequest(&request, ctx.Value(middleware.USER_GROUP_KEY).(middleware.UserGroup))
	if err != nil {
		return common.RequestErrorResponse(
			http.StatusBadRequest,
			"Invalid request path parameter",
			logger,
		)
	}

	application, err := applicationRepositoryService.GetByID(context.TODO(), applicationId)
	if err != nil {
		return utils.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to get application by id: %v", err), logger)
	}

	jsonApplication, err := json.Marshal(*application)
	if err != nil {
		return utils.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal scan output items: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonApplication),
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN, middleware.USER))
}

func processRequest(request *events.APIGatewayV2HTTPRequest, group middleware.UserGroup) (string, error) {
	switch group {
	case middleware.ADMIN:
		id, ok := request.PathParameters["id"]
		if !ok {
			return "", fmt.Errorf("missing parameter 'id'")
		}
		id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")
		return id, nil
	case middleware.USER:
		cognitoID, ok := request.RequestContext.Authorizer.JWT.Claims["sub"]
		if !ok {
			return "", fmt.Errorf("missing sub in JWT claims")
		}
		id, err := applicationRepositoryService.GetApplicationIDByCognitoID(cognitoID)
		if err != nil {
			return "", fmt.Errorf("failed to get Application ID by CognitoID")
		}
		return id, nil
	default:
		return "", errors.New("process request failed: Invalid user group")
	}
}
