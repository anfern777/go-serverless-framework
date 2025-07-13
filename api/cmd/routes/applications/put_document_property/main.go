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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var (
	err                          error
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]

	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	logger    = implementations.NewSlogLogger()
)

type RequestBody struct {
	Property string `json:"property"`
	Value    string `json:"value"`
}

func init() {
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName)
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, property, propertyType, documentType, atValue, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to process path parameters from request: %s", err), logger)
	}

	err = documentRepositoryService.UpdateProperty(applicationId, models.CreateDocumentSK(*documentType), property, propertyType, *atValue)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to update property '%s': %s", property, err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

// returns in order: applicationId, updateExpressionKey, updateExpressionType, documentType, attributeValue, error
func processRequest(request *events.APIGatewayV2HTTPRequest) (string, string, string, *models.DocumentType, *types.AttributeValue, error) {
	id, exists := request.PathParameters["id"]
	if !exists || id == "" {
		return "", "", "", nil, nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}
	id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")

	documentType, exists := request.PathParameters["document-type"]
	if !exists || documentType == "" {
		return "", "", "", nil, nil, errors.New("failed to get path parameter 'document-type'")
	}

	var requestBody RequestBody
	err = json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return "", "", "", nil, nil, errors.New("failed to get request body")
	}

	var updateExpressionKey, updateExpressionType string
	var value any
	var atValue types.AttributeValue
	switch requestBody.Property {
	case "notes":
		updateExpressionKey = "Notes"
		updateExpressionType = "s"
		value = requestBody.Value
		atValue = &types.AttributeValueMemberS{Value: value.(string)}
	default:
		return "", "", "", nil, nil, fmt.Errorf("invalid propterty '%s'", requestBody.Property)
	}

	return id, updateExpressionKey, updateExpressionType, common.PtrTo(models.DocumentType(documentType)), &atValue, nil
}
