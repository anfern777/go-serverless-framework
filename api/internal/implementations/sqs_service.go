package implementations

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

// implements services.MessageQueuing
type SQSMessageQueuing[T any] struct {
	client   SQSClientAPI
	queueUrl string
}

type SQSClientAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	SendMessageBatch(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
}

func NewSQSService[T any](queueUrl string) (*SQSMessageQueuing[T], error) {
	client, err := getSQSClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get SQS client; reason: %w", err)
	}
	service := &SQSMessageQueuing[T]{
		client:   client,
		queueUrl: queueUrl,
	}
	return service, nil
}

func (s *SQSMessageQueuing[T]) SendMessage(params *T) error {
	var zeroT T
	switch any(zeroT).(type) {
	case services.EmailerInputParams:
		params := any(params).(*services.EmailerInputParams)
		if params == nil || params.Template == "" || params.Destination == "" || params.Source == "" || params.Subject == "" {
			return fmt.Errorf("params cannot be nil or empty")
		}
	}
	jsonMessage, err := json.Marshal(*params)
	if err != nil {
		return fmt.Errorf("failed to marshal email; reason: %w", err)
	}
	_, err = s.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
		MessageBody: common.PtrTo(string(jsonMessage)),
		QueueUrl:    &s.queueUrl,
	})
	if err != nil {
		return fmt.Errorf("failed to send message to sqs; reason: %w", err)
	}
	return nil
}

// func (s *SQSMessageQueuing[T]) SendMessageBatch(params []T) error {
// 	var zeroT T
// 	switch any(zeroT).(type) {
// 	case services.EmailerInputParams:
// 		params := any(params).([]services.EmailerInputParams)
// 		for _, param := range params {
// 			if param.Template == "" || param.Destination == "" || param.Source == "" || param.Subject == "" {
// 				return fmt.Errorf("params cannot be nil or empty")
// 			}
// 		}
// 	}
// 	jsonMessage, err := json.Marshal(params)
// 	if err != nil {
// 		return fmt.Errorf("failed to marshal email; reason: %w", err)
// 	}
// 	_, err = s.client.SendMessage(context.TODO(), &sqs.SendMessageInput{
// 		MessageBody: common.PtrTo(string(jsonMessage)),
// 		QueueUrl:    &s.queueUrl,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to send message to sqs; reason: %w", err)
// 	}
// 	return nil
// }

func getSQSClient() (*sqs.Client, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("utils: failed to load s3 config: %v", err)
	}

	return sqs.NewFromConfig(cfg), nil
}
