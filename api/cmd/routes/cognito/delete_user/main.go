package main

import (
	"context"
	"encoding/json"
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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RequestBody struct {
	Email string `json:"email"`
	ID    string `json:"id"`
}

var (
	err                          error
	logger                       services.Logger
	authService                  services.AuthProvider
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService repository.BaseRepository[models.Application]

	userPoolId = os.Getenv("USERPOOL_ID")
	tableName  = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
)

func init() {
	authService, err = implementations.NewCognitoAuthProvider()
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize Cognito service: %v", err)
	}
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName)
}

func handleRequest(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	email, applicationId, err := processRequest(req)
	if err != nil {
		common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err), logger)
	}

	if err = authService.DeleteUser(services.AuthProviderAccessParams{
		UserpoolID: &userPoolId,
		Email:      &email,
	}); err != nil {
		common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to create user in cognito: %v", err), logger)
	}

	// update application data in dynamoDb
	err = applicationRepositoryService.UpdateProperty(applicationId, "CognitoID", "s", &types.AttributeValueMemberS{Value: ""})
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to update property '%s': %s", "CognitoID", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request events.APIGatewayV2HTTPRequest) (string, string, error) {
	var requestBody RequestBody

	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return "", "", fmt.Errorf("failed to parse request body: %w", err)
	}
	id := strings.Join([]string{common.PK_APP_PREFIX, "#", requestBody.ID}, "")

	return requestBody.Email, id, nil
}
