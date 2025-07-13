package main

import (
	"context"
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/implementations"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

var (
	err error

	fsService  services.FileStorageService
	emailer    services.Emailer
	logger     services.Logger
	bucketName = os.Getenv(common.LAMBDAENV_BUCKET_NAME)
)

func init() {
	logger = implementations.NewSlogLogger()
	fsService, err = implementations.NewS3FileStorageService(bucketName)
	if err != nil {
		log.Fatalf("Failed to initialize file storage service")
	}
	emailer, err = implementations.NewSesEmailerImpl(implementations.WithFsService(fsService))
	if err != nil {
		log.Fatalf("Failed to initialize email service")
	}
}

func handleRequest(ctx context.Context, request events.SQSEvent) (*events.SQSEventResponse, error) {
	var wg sync.WaitGroup
	var batchItemFailuresCH = make(chan events.SQSBatchItemFailure, len(request.Records))

	wg.Add(len(request.Records))
	for _, record := range request.Records {
		go processMessage(&wg, record, batchItemFailuresCH)
	}
	wg.Wait()
	close(batchItemFailuresCH)

	var batchItemFailures []events.SQSBatchItemFailure
	for itemFailure := range batchItemFailuresCH {
		batchItemFailures = append(batchItemFailures, itemFailure)
	}

	return &events.SQSEventResponse{
		BatchItemFailures: batchItemFailures,
	}, nil
}

func main() {
	lambda.Start(handleRequest)
}

func processMessage(wg *sync.WaitGroup, record events.SQSMessage, batchItemFailuresCH chan<- events.SQSBatchItemFailure) {
	var emailMetadata services.EmailerInputParams
	defer wg.Done()

	err := json.Unmarshal([]byte(record.Body), &emailMetadata)
	if err != nil {
		batchItemFailuresCH <- events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
		logger.Log(
			"Failed to process message",
			slog.LevelError,
			"MessageID", record.MessageId,
			"Error", err,
		)
		return
	}

	err = emailer.Email(emailMetadata)
	if err != nil {
		batchItemFailuresCH <- events.SQSBatchItemFailure{ItemIdentifier: record.MessageId}
		logger.Log(
			"Failed to process message",
			slog.LevelError,
			"MessageID", record.MessageId,
			"Error", err,
		)
		return
	}
}
