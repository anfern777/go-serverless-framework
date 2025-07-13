package implementations

import (
	"context"
	"errors"
	"fmt"

	"github.com/anfern777/go-serverless-framework-api/internal/common"
	"github.com/anfern777/go-serverless-framework-api/internal/middleware"
	"github.com/anfern777/go-serverless-framework-api/internal/services"

	cognito "github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider"
	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/types"
	"github.com/aws/aws-sdk-go/aws"
)

type CognitoAuthProvider struct {
	client CognitoClient
}

type CognitoClient interface {
	AdminCreateUser(ctx context.Context, params *cognito.AdminCreateUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminCreateUserOutput, error)
	AdminAddUserToGroup(ctx context.Context, params *cognito.AdminAddUserToGroupInput, optFns ...func(*cognito.Options)) (*cognito.AdminAddUserToGroupOutput, error)
	AdminDeleteUser(ctx context.Context, params *cognito.AdminDeleteUserInput, optFns ...func(*cognito.Options)) (*cognito.AdminDeleteUserOutput, error)
}

func NewCognitoAuthProvider() (*CognitoAuthProvider, error) {
	client, err := common.GetCognitoClient()
	if err != nil {
		return nil, fmt.Errorf("failed to get cognito client")
	}
	return &CognitoAuthProvider{
		client: client,
	}, nil
}

func (c *CognitoAuthProvider) CreateUser(authAccessParams services.AuthProviderAccessParams) (any, error) {
	output, err := c.client.AdminCreateUser(context.TODO(), &cognito.AdminCreateUserInput{
		UserPoolId:             authAccessParams.UserpoolID,
		Username:               authAccessParams.Email,
		DesiredDeliveryMediums: []types.DeliveryMediumType{types.DeliveryMediumTypeEmail},
	})
	if err != nil {
		var userExists *types.UsernameExistsException
		if errors.As(err, &userExists) {
			// Specific error for user already existing
			return nil, fmt.Errorf("user %s already exists in the user pool: %w", *authAccessParams.Email, err)
		} else {
			// Generic error for other Cognito issues, preserving the original error for logging
			return nil, fmt.Errorf("user creation for %s encountered an unexpected issue. Please check service configuration or logs for details: %w", *authAccessParams.Email, err)
		}
	}

	_, err = c.client.AdminAddUserToGroup(context.TODO(), &cognito.AdminAddUserToGroupInput{
		GroupName:  aws.String(string(middleware.USER)),
		UserPoolId: authAccessParams.UserpoolID,
		Username:   output.User.Username,
	})
	if err != nil {
		return "", fmt.Errorf("failed to add user to user group: %w", err)
	}

	cognitoID, ok := getSubAttributeValue(output.User.Attributes)
	if !ok || cognitoID == "" {
		return "", fmt.Errorf("failed to retrieve Cognito 'sub' ID after user creation")
	}

	return cognitoID, nil
}

func (c *CognitoAuthProvider) DeleteUser(authAccessParams services.AuthProviderAccessParams) error {
	_, err := c.client.AdminDeleteUser(context.TODO(), &cognito.AdminDeleteUserInput{
		UserPoolId: authAccessParams.UserpoolID,
		Username:   authAccessParams.Email,
	})
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}

func getSubAttributeValue(attributes []types.AttributeType) (string, bool) {
	if attributes == nil {
		return "", false
	}
	for _, attr := range attributes {
		if attr.Name != nil && *attr.Name == "sub" {
			if attr.Value != nil {
				return *attr.Value, true
			}
			return "", false
		}
	}
	return "", false
}
