package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/implementations"
	"github.com/anfern777/go-serverless-framework-api/internal/models"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	err          error
	queueService services.MessageQueuing[services.EmailerInputParams]

	logger           = implementations.NewSlogLogger()
	officeEmail      = os.Getenv("OFFICE_EMAIL")
	applicationEmail = os.Getenv("APPLICATION_EMAIL")
	queueURL         = os.Getenv(common.LAMBDA_ENV_SQS_QUEUE_URL)

	//go:embed email-template/contact-request.html
	template string
)

func init() {
	queueService, err = implementations.NewSQSService[services.EmailerInputParams](queueURL)
	if err != nil {
		log.Fatalf("Failed to initialize queue service: %v", err)
	}
}

func handleRequest(ctx context.Context, request events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	var contactInfo models.Contact
	err := json.Unmarshal([]byte(request.Body), &contactInfo)
	if err != nil {
		return common.RequestErrorResponse(http.StatusBadRequest, fmt.Sprintf("Failed to marshal input: %v", err), logger)
	}

	if err = queueService.SendMessage(
		&services.EmailerInputParams{
			Template:    template,
			Subject:     "go-serverless-framework: Contact request received",
			Source:      applicationEmail,
			Destination: officeEmail,
			Data: map[string]string{
				"Name":    contactInfo.Name,
				"Email":   contactInfo.Email,
				"Message": contactInfo.Message,
			},
		}); err != nil {
		return common.RequestErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to send email: %v", err), logger)
	}

	return &events.APIGatewayV2HTTPResponse{
		StatusCode: http.StatusNoContent,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}
