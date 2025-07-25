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
	logger                       services.Logger

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
	messageRepositoryService = repository.NewChildRepositoryImpl[models.Message](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, cognitoID, err := processRequest(&request, ctx.Value(middleware.USER_GROUP_KEY).(middleware.UserGroup))
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Invalid request: %v", err), logger)
	}

	if applicationId == "" {
		applicationId, err = applicationRepositoryService.GetApplicationIDByCognitoID(cognitoID)
		if err != nil {
			return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to query table: %v", err), logger)
		}
	}

	result, err := messageRepositoryService.GetAllByParentPK(applicationId)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to get messages: %v", err), logger)
	}

	resultJson, err := json.Marshal(result)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal query output items: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(resultJson),
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN, middleware.USER))
}

func processRequest(request *events.APIGatewayV2HTTPRequest, group middleware.UserGroup) (string, string, error) {
	switch group {
	case middleware.ADMIN:
		id, ok := request.PathParameters["id"]
		if !ok {
			return "", "", fmt.Errorf("missing parameter 'id'")
		}
		id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")
		return id, "", nil
	case middleware.USER:
		cognitoID, ok := request.RequestContext.Authorizer.JWT.Claims["sub"]
		if !ok {
			return "", "", fmt.Errorf("missing cognitoID in JWT claims")
		}
		return "", cognitoID, nil
	}

	return "", "", errors.New("failed to get application cognitoID or ID from request - missing or empty 'id'|'cognitoID' path parameter")
}
