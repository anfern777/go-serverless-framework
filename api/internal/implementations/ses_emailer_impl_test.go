package implementations

import (
	"context"
	_ "embed"
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/services"

	"github.com/aws/aws-sdk-go-v2/service/ses"
)

//go:embed fixtures/test-email.html
var testEmailTemplate string

type MockSesClientAPI struct {
	SendRawEmailFunc  func(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error)
	SendRawEmailCalls []struct {
		Data []byte
	}
}

func (m *MockSesClientAPI) SendRawEmail(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error) {
	m.SendRawEmailCalls = append(m.SendRawEmailCalls, struct{ Data []byte }{
		Data: params.RawMessage.Data,
	})
	return m.SendRawEmailFunc(ctx, params, optFns...)
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name                         string
		input                        services.EmailerInputParams
		sendRawEmailFunc             func(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error)
		expectedSendRawFunctionCalls int
	}{
		{
			name: "happy path",
			input: services.EmailerInputParams{
				Template:    testEmailTemplate,
				Subject:     "Test Email",
				Source:      "source@email.com",
				Destination: "destination@email.com",
				Data: map[string]string{
					"Name": "Upwigo",
				},
				Attachments: nil,
			},
			sendRawEmailFunc: func(ctx context.Context, params *ses.SendRawEmailInput, optFns ...func(*ses.Options)) (*ses.SendRawEmailOutput, error) {
				return &ses.SendRawEmailOutput{}, nil
			},
			expectedSendRawFunctionCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockSesClientAPI{
				SendRawEmailFunc: tt.sendRawEmailFunc,
			}
			sesEmail := &SesEmailerImpl{
				client: mockClient,
			}
			err := sesEmail.Email(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}
			callsCount := len(mockClient.SendRawEmailCalls)
			if tt.expectedSendRawFunctionCalls != callsCount {
				t.Errorf("Expected SendRawEmail to be called %d times, but got called %d times", tt.expectedSendRawFunctionCalls, callsCount)
			}
		})
	}
}
