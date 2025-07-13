package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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

	logger    = implementations.NewSlogLogger()
	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	indexName = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
)

func init() {
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	applicationRepositoryService = repository.NewApplicationsRepositoryImpl(databaseService, tableName, repository.WithGSIIndex[models.Application](indexName))
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	startDate, endDate, err := processRequest(request)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err), logger)
	}

	items, err := applicationRepositoryService.GetBetweenDates(context.TODO(), startDate, endDate, common.PK_APP_PREFIX)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to query items: %v", err), logger)
	}

	jsonItems, err := json.Marshal(items)
	if err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal query output items: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(jsonItems),
	}, nil
}

func main() {
	lambda.Start(middleware.AuthorizationMiddleware(handleRequest, logger, middleware.ADMIN))
}

func processRequest(request events.APIGatewayV2HTTPRequest) (string, string, error) {
	startDate, exists1 := request.PathParameters["startDate"]
	endDate, exists2 := request.PathParameters["endDate"]
	if !exists1 || !exists2 || startDate == "" {
		sixMonthsAgo := time.Now().AddDate(0, -6, 0)
		startDate = sixMonthsAgo.Format(time.RFC3339)
		endDate = time.Now().Format(time.RFC3339)
		return startDate, endDate, nil
	}
	return startDate, endDate, nil
}
