package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	utils "github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/implementations"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/repository"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	err             error
	logger          services.Logger
	databaseService services.DatabaseClientAPI
	postsRepository *repository.PostsRepositoryImpl

	tableName = os.Getenv(common.LAMBDA_ENV_TABLE_NAME)
	index     = os.Getenv(common.LAMBDA_ENV_GSI_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()
	databaseService, err = implementations.NewDynamodbDatabaseService(tableName)
	if err != nil {
		log.Fatalf("failed to initialize database service - execution stopped")
	}
	postsRepository = repository.NewPostsRepositoryImpl(databaseService, tableName, repository.WithBaseAssociationsIndex[models.Post](index))
}

// supportedLanguages map to check against.
// Using common.LanguageCode for consistency with the rest of the project.
var supportedLanguages = map[common.LanguageCode]bool{
	common.EN: true,
	common.DE: true,
	common.TL: true,
	common.AT: true,
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	languageStr, exists := request.QueryStringParameters[common.QUERY_PARAM_LANGUAGE]
	if !exists {
		return utils.RequestErrorResponse(http.StatusBadRequest, "Failed to get language from request: 'lang' query parameter is missing", logger)
	}

	language := common.LanguageCode(languageStr)
	if _, ok := supportedLanguages[language]; !ok {
		return utils.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf(`"Unsupported language code: %s"}`, languageStr), logger)
	}

	items, err := postsRepository.GetByLanguage(string(language))
	if err != nil {
		return utils.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to query posts by language %v", err), logger)
	}

	jsonItems, err := json.Marshal(items)
	if err != nil {
		return utils.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to marshal query output items: %v", err), logger)
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
	lambda.Start(handleRequest)
}
