package middleware

import (
	"context"
	"log/slog"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

type MockLogger struct{}

func (ml *MockLogger) Log(message string, level slog.Level, args ...any) error {
	return nil
}

type NextSpy struct {
	wasCalled       bool
	capturedContext context.Context
}

func (ns *NextSpy) serve(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error) {
	ns.wasCalled = true
	ns.capturedContext = ctx
	return &events.APIGatewayV2HTTPResponse{}, nil
}

func TestAuthorizationMiddleware(t *testing.T) {
	mockLogger := &MockLogger{}

	mockRequest := func(groupString string) events.APIGatewayV2HTTPRequest {
		return events.APIGatewayV2HTTPRequest{
			RequestContext: events.APIGatewayV2HTTPRequestContext{
				Authorizer: &events.APIGatewayV2HTTPRequestContextAuthorizerDescription{
					JWT: &events.APIGatewayV2HTTPRequestContextAuthorizerJWTDescription{
						Claims: map[string]string{
							"cognito:groups": groupString,
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name          string
		allowedGroups []UserGroup
		groupString   string

		expectedGroupFromContext UserGroup
		expectError              bool
	}{
		{
			name: "happy path",
			allowedGroups: []UserGroup{
				ADMIN,
			},
			groupString: "[admin]",

			expectedGroupFromContext: ADMIN,
			expectError:              false,
		},
		{
			name: "malformed group claims",
			allowedGroups: []UserGroup{
				ADMIN,
			},
			groupString: "[admin",

			expectedGroupFromContext: "",
			expectError:              true,
		},
		{
			name: "invalid user",
			allowedGroups: []UserGroup{
				ADMIN,
			},
			groupString: "[invaliduser]",

			expectedGroupFromContext: "",
			expectError:              true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextSpy := &NextSpy{}
			authMiddleware := AuthorizationMiddleware(nextSpy.serve, mockLogger, tt.allowedGroups...)
			_, err := authMiddleware(context.Background(), mockRequest(tt.groupString))

			if err != nil {
				if !tt.expectError {
					t.Errorf("Expected no error, but got %v", err)
				}
				if nextSpy.wasCalled {
					t.Errorf("Expected no call for next, but next got called")
				}
			}
			if err == nil {
				if tt.expectError {
					t.Error("Expected error, but got no error")
				}
				if !nextSpy.wasCalled {
					t.Errorf("Expected next to be called, but got no call")
				}
				capturedGroup := nextSpy.capturedContext.Value(USER_GROUP_KEY).(UserGroup)
				if nextSpy.capturedContext.Value(USER_GROUP_KEY).(UserGroup) != tt.expectedGroupFromContext {
					t.Errorf("Expected group %s, but got group %s", string(tt.expectedGroupFromContext), string(capturedGroup))
				}
			}
		})
	}
}

func TestGetUserGroupFromJWTClaims(t *testing.T) {
	tests := []struct {
		name   string
		claims map[string]string

		expectError bool
	}{
		{
			name: "happy path",
			claims: map[string]string{
				"cognito:groups": "[admin]",
			},
			expectError: false,
		},
		{
			name: "invalid group",
			claims: map[string]string{
				"cognito:groups": "[blabla]",
			},
			expectError: true,
		},
		{
			name: "invalid format",
			claims: map[string]string{
				"cognito:groups": "{admin}",
			},
			expectError: true,
		},
		{
			name:        "claims is nil - invalid",
			claims:      nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := GetUserGroupFromJWTClaims(tt.claims)
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error, but got %v", err)
			}
			if tt.expectError && err == nil {
				t.Errorf("Expected error, but got no error")
			}
		})
	}
}
