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
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type RequestBody struct {
	Property string          `json:"property"`
	Value    json.RawMessage `json:"value"`
}

var (
	err                          error
	databaseService              services.DatabaseClientAPI
	applicationRepositoryService *repository.ApplicationsRepositoryImpl
	documentRepositoryService    *repository.ChildRepositoryImpl[models.Document]
	tableName                    = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	logger                       = implementations.NewSlogLogger()
)

func init() {
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName)
	documentRepositoryService = repository.NewChildRepositoryImpl[models.Document](databaseService, tableName)
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	applicationId, property, propertyType, atValue, err := processRequest(&request)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to process path parameters from request: %s", err), logger)
	}

	// update application data in dynamoDb
	err = applicationRepositoryService.UpdateProperty(applicationId, property, propertyType, *atValue)
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

// returns the application id and the update expression key for dynamodb update operation
func processRequest(request *events.APIGatewayV2HTTPRequest) (string, string, string, *types.AttributeValue, error) {
	id, exists := request.PathParameters["id"]
	if !exists || id == "" {
		return "", "", "", nil, errors.New("failed to get application ID from request - missing or empty 'id' path parameter")
	}
	id = strings.Join([]string{common.PK_APP_PREFIX, "#", id}, "")

	var requestBody RequestBody
	err := json.Unmarshal([]byte(request.Body), &requestBody)
	if err != nil {
		return "", "", "", nil, errors.New("failed to get request body")
	}

	var updateExpressionKey, updateExpressionType string
	var atValue types.AttributeValue
	switch requestBody.Property {
	case "analysed":
		updateExpressionKey = "Analysed"
		updateExpressionType = "b"
		var boolValue bool

		if err := json.Unmarshal(requestBody.Value, &boolValue); err != nil {
			return "", "", "", nil, fmt.Errorf("failed to unmarshal boolean value for property '%s': %w", requestBody.Property, err)
		}

		atValue = &types.AttributeValueMemberBOOL{Value: boolValue}
	case "pre-screening":
		updateExpressionKey = "PreScreeningStatus"
		updateExpressionType = "s"
		var stringValue string

		if err := json.Unmarshal(requestBody.Value, &stringValue); err != nil {
			return "", "", "", nil, fmt.Errorf("failed to unmarshal string value for property '%s': %w", requestBody.Property, err)
		}

		atValue = &types.AttributeValueMemberS{Value: stringValue}
	case "german-training-info":
		updateExpressionKey = "GermanTrainingInfo"
		updateExpressionType = "m"

		var germanTrainingInfoValue models.GermanTrainingInfo
		if err := json.Unmarshal(requestBody.Value, &germanTrainingInfoValue); err != nil {
			return "", "", "", nil, fmt.Errorf("failed to unmarshal 'value' for property '%s' into %T: %w", requestBody.Property, germanTrainingInfoValue, err)
		}

		valueMapAv, err := attributevalue.MarshalMap(germanTrainingInfoValue)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to marshal map value for '%s': %w", requestBody.Property, err)
		}
		atValue = &types.AttributeValueMemberM{Value: valueMapAv}
	case "employer-info":
		updateExpressionKey = "EmployerInfo"
		updateExpressionType = "m"

		var employerInfoValue models.EmployerInfo
		if err := json.Unmarshal(requestBody.Value, &employerInfoValue); err != nil {
			return "", "", "", nil, fmt.Errorf("failed to unmarshal 'value' for property '%s' into %T: %w", requestBody.Property, employerInfoValue, err)
		}

		valueMapAv, err := attributevalue.MarshalMap(employerInfoValue)
		if err != nil {
			return "", "", "", nil, fmt.Errorf("failed to marshal map value for '%s': %w", requestBody.Property, err)
		}
		atValue = &types.AttributeValueMemberM{Value: valueMapAv}
	default:
		return "", "", "", nil, fmt.Errorf("invalid property '%s'", requestBody.Property)
	}

	return id, updateExpressionKey, updateExpressionType, &atValue, nil
}
