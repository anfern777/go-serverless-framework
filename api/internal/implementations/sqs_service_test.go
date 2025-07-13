package implementations

import (
	"context"
	_ "embed"
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type MockSQSClientAPI struct {
	SendMessageFunc      func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
	SendMessageFuncCalls []struct {
		params *sqs.SendMessageInput
	}

	SendBatchMessageFunc      func(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error)
	SendBatchMessageFuncCalls []struct {
		params *sqs.SendMessageBatchInput
	}
}

func (m *MockSQSClientAPI) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.SendMessageFuncCalls = append(
		m.SendMessageFuncCalls,
		struct {
			params *sqs.SendMessageInput
		}{
			params: params,
		})
	return m.SendMessageFunc(ctx, params)
}

func (m *MockSQSClientAPI) SendMessageBatch(ctx context.Context, params *sqs.SendMessageBatchInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageBatchOutput, error) {
	m.SendBatchMessageFuncCalls = append(
		m.SendBatchMessageFuncCalls,
		struct {
			params *sqs.SendMessageBatchInput
		}{
			params: params,
		})
	return m.SendBatchMessageFunc(ctx, params)
}

func TestSendMessage(t *testing.T) {
	tests := []struct {
		name                         string
		input                        *services.EmailerInputParams
		expectedSendMessageFuncCalls int
		expectedError                bool
		expectedParams               *sqs.SendMessageInput
	}{
		{
			name: "happy path",
			input: &services.EmailerInputParams{
				Subject:     "test-subject",
				Source:      "test@email.com",
				Destination: "test-dest@email.com",
				Data: map[string]string{
					"Name": "John Doe",
				},
				Attachments: nil,
				Template:    "test-template",
			},
			expectedParams: &sqs.SendMessageInput{
				MessageBody: common.PtrTo(`{"Template":"test-template","Subject":"test-subject","Source":"test@email.com","Destination":"test-dest@email.com","Data":{"Name":"John Doe"},"Attachments":null}`),
				QueueUrl:    common.PtrTo("test-queue-url"),
			},
			expectedSendMessageFuncCalls: 1,
			expectedError:                false,
		},
		{
			name:                         "error sending message",
			input:                        nil,
			expectedSendMessageFuncCalls: 0,
			expectedError:                true,
		},
		{
			name:                         "empty input",
			input:                        &services.EmailerInputParams{},
			expectedSendMessageFuncCalls: 0,
			expectedError:                true,
		},
	}

	mockSQSClientAPI := &MockSQSClientAPI{
		SendMessageFunc: func(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
			return &sqs.SendMessageOutput{}, nil
		},
	}
	mockSQS := &SQSMessageQueuing[services.EmailerInputParams]{
		client:   mockSQSClientAPI,
		queueUrl: "test-queue-url",
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				mockSQSClientAPI.SendMessageFuncCalls = []struct {
					params *sqs.SendMessageInput
				}{}
			}()

			err := mockSQS.SendMessage(tt.input)
			if err != nil {
				if tt.expectedError == false {
					t.Errorf("expected no error, but got %v", err)
				}
			} else {
				if tt.expectedSendMessageFuncCalls != len(mockSQSClientAPI.SendMessageFuncCalls) {
					t.Errorf("expected %d calls, but got %d", tt.expectedSendMessageFuncCalls, len(mockSQSClientAPI.SendMessageFuncCalls))
				}
				if *tt.expectedParams.MessageBody != *mockSQSClientAPI.SendMessageFuncCalls[0].params.MessageBody {
					t.Errorf("expected params \n%v\n,but got \n%v\n", *tt.expectedParams.MessageBody, *mockSQSClientAPI.SendMessageFuncCalls[0].params.MessageBody)
				}
				if tt.expectedParams.QueueUrl != mockSQSClientAPI.SendMessageFuncCalls[0].params.QueueUrl {

				}
			}
		})
	}
}
