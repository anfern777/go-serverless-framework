package implementations

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/anfern777/go-serverless-framework-api/internal/services"

	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/smithy-go/middleware"
)

type MockCognitoClient struct {
	AdminCreateUserFunc       func(ctx context.Context, params *cognito.AdminCreateUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminCreateUserOutput, error)
	AdminCreateUserCallsCount int

	AdminAddUserToGroupFunc       func(ctx context.Context, params *cognito.AdminAddUserToGroupInput, optFns ...func(*cognito.Options)) (*cognito.AdminAddUserToGroupOutput, error)
	AdminAddUserToGroupCallsCount int

	AdminDeleteUserFunc       func(ctx context.Context, params *cognito.AdminDeleteUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminDeleteUserOutput, error)
	AdminDeleteUserCallsCount int
}

func (mcc *MockCognitoClient) AdminCreateUser(ctx context.Context, params *cognito.AdminCreateUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminCreateUserOutput, error) {
	mcc.AdminCreateUserCallsCount++
	return mcc.AdminCreateUserFunc(ctx, params)
}

func (mcc *MockCognitoClient) AdminAddUserToGroup(ctx context.Context, params *cognito.AdminAddUserToGroupInput, optFns ...func(*cognito.Options)) (*cognito.AdminAddUserToGroupOutput, error) {
	mcc.AdminAddUserToGroupCallsCount++
	return mcc.AdminAddUserToGroupFunc(ctx, params)
}

func (mcc *MockCognitoClient) AdminDeleteUser(ctx context.Context, params *cognito.AdminDeleteUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminDeleteUserOutput, error) {
	mcc.AdminDeleteUserCallsCount++
	return mcc.AdminDeleteUserFunc(ctx, params)
}

func TestCreateUser(t *testing.T) {
	mockClient := &MockCognitoClient{}

	testsForAdminCreateUserSuccess := []struct {
		name                             string
		userPoolId                       string
		email                            string
		expectedCognitoId                string
		expectedAddUserCallsCount        int
		expectedAddUserToGroupCallsCount int
		expectedErrorCreateUser          error
		expectedErrorAddUserToGroup      error
	}{
		{
			name:       "happy path",
			userPoolId: "eu-west-1_12345",
			email:      "test@example.com",

			expectedCognitoId:                "a1b2c3d4-e5f6-7890-1234-567890abcdef",
			expectedAddUserCallsCount:        1,
			expectedErrorCreateUser:          nil,
			expectedErrorAddUserToGroup:      nil,
			expectedAddUserToGroupCallsCount: 1,
		},
		// {
		// 	name:       "invalid user name",
		// 	userPoolId: "eu-west-1_12345",
		// 	email:      "",

		// 	expectedCognitoId:                "a1b2c3d4-e5f6-7890-1234-567890abcdef",
		// 	expectedAddUserCallsCount:        1,
		// 	expectedAddUserToGroupCallsCount: 0,
		// 	expectedErrorCreateUser:          nil,
		// 	expectedErrorAddUserToGroup:      nil,
		// },
		// {
		// 	name:       "invalid user pool id",
		// 	userPoolId: "",
		// 	email:      "test@example.com",

		// 	expectedCognitoId:                "a1b2c3d4-e5f6-7890-1234-567890abcdef",
		// 	expectedAddUserCallsCount:        1,
		// 	expectedAddUserToGroupCallsCount: 0,
		// 	expectedErrorCreateUser:          &types.UsernameExistsException{},
		// 	expectedErrorAddUserToGroup:      nil,
		// },
	}

	for _, tt := range testsForAdminCreateUserSuccess {
		t.Run(tt.name, func(t *testing.T) {
			mockClient.AdminCreateUserFunc = func(ctx context.Context, params *cognito.AdminCreateUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminCreateUserOutput, error) {
				return &cognito.AdminCreateUserOutput{
					User: &types.UserType{
						Username: params.Username,
						Attributes: []types.AttributeType{
							{
								Name:  aws.String("sub"),
								Value: aws.String(tt.expectedCognitoId),
							},
							{
								Name:  aws.String("email"),
								Value: aws.String(tt.email),
							},
						},
					},
				}, nil
			}
			mockClient.AdminAddUserToGroupFunc = func(ctx context.Context, params *cognito.AdminAddUserToGroupInput, optFns ...func(*cognito.Options)) (*cognito.AdminAddUserToGroupOutput, error) {
				return &cognito.AdminAddUserToGroupOutput{
					ResultMetadata: middleware.Metadata{},
				}, nil
			}
			cognitoAuthProvider := &CognitoAuthProvider{
				client: mockClient,
			}
			_, err := cognitoAuthProvider.CreateUser(services.AuthProviderAccessParams{
				UserpoolID: &tt.userPoolId,
				Email:      &tt.email,
			})
			if err == nil && (errors.Is(err, tt.expectedErrorCreateUser) && tt.expectedErrorCreateUser != nil) {
				t.Errorf("expected error of type %v, but got no error", reflect.TypeOf(tt.expectedErrorCreateUser))
			}
			if err != nil && tt.expectedErrorCreateUser == nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if err != nil && tt.expectedErrorAddUserToGroup == nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if mockClient.AdminCreateUserCallsCount != tt.expectedAddUserCallsCount {
				t.Errorf("expected 1 call, but got %d", mockClient.AdminCreateUserCallsCount)
			}
			if mockClient.AdminAddUserToGroupCallsCount != tt.expectedAddUserToGroupCallsCount {
				t.Errorf("expected 1 call, but got %d", mockClient.AdminAddUserToGroupCallsCount)
			}
			defer func() {
				mockClient = &MockCognitoClient{}
			}()
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name                         string
		expectedDeleteuserCallsCount int
		userPoolId                   string
		email                        string
		expectedError                error
	}{
		{
			name:                         "happy path",
			expectedDeleteuserCallsCount: 1,
			userPoolId:                   "eu-west-1_12345",
			email:                        "test@example.com",
			expectedError:                nil,
		},
	}

	mockClient := &MockCognitoClient{
		AdminDeleteUserFunc: func(ctx context.Context, params *cognito.AdminDeleteUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminDeleteUserOutput, error) {
			return &cognito.AdminDeleteUserOutput{
				ResultMetadata: middleware.Metadata{},
			}, nil
		},
	}
	cognitoAuthProvider := &CognitoAuthProvider{
		client: mockClient,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := cognitoAuthProvider.DeleteUser(services.AuthProviderAccessParams{
				Email:      &tt.email,
				UserpoolID: &tt.email,
			})
			if err != nil && tt.expectedError == nil {
				t.Errorf("expected no error, but got %v", err)
			}
			if mockClient.AdminDeleteUserCallsCount != tt.expectedDeleteuserCallsCount {
				t.Errorf("expected %d calls, but got %d calls", tt.expectedDeleteuserCallsCount, mockClient.AdminDeleteUserCallsCount)
			}
			defer func() {
				mockClient = &MockCognitoClient{}
			}()
		})
	}
}

func TestGetSubAttributeValue(t *testing.T) {
	tests := []struct {
		name       string
		attributes []types.AttributeType

		expectedResult        string
		expectedBooleanResult bool
	}{
		{
			name: "happy path",
			attributes: []types.AttributeType{
				{
					Name:  aws.String("sub"),
					Value: aws.String("a1b2c3d4-e5f6-7890-1234-567890abcdef"),
				},
			},
			expectedResult:        "a1b2c3d4-e5f6-7890-1234-567890abcdef",
			expectedBooleanResult: true,
		},
		{
			name: "sub does not exist in attributes",
			attributes: []types.AttributeType{
				{
					Name:  aws.String("email"),
					Value: aws.String("test@example.com"),
				},
			},
			expectedResult:        "",
			expectedBooleanResult: false,
		},
		{
			name: "value of sub attribute is nil",
			attributes: []types.AttributeType{
				{
					Name:  aws.String("sub"),
					Value: nil,
				},
			},
			expectedResult:        "",
			expectedBooleanResult: false,
		},
		{
			name:                  "attributes is empty",
			attributes:            []types.AttributeType{},
			expectedResult:        "",
			expectedBooleanResult: false,
		},
		{
			name:                  "attributes is nil",
			attributes:            nil,
			expectedResult:        "",
			expectedBooleanResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := getSubAttributeValue(tt.attributes)
			if !ok && tt.expectedBooleanResult {
				t.Errorf("expected ok, but got %t", ok)
			}
			if ok && !tt.expectedBooleanResult {
				t.Errorf("expected not ok, but got %t", ok)
			}
			if ok && result != tt.expectedResult {
				t.Errorf("expected %s, but got %s", tt.expectedResult, result)
			}
			if tt.attributes == nil && (result != "" || ok) {
				t.Errorf("expected empty result and not ok, but got result: %s and ok: %t", result, ok)
			}
		})
	}
}
