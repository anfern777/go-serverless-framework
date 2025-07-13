package backend

import (
	"fmt"

	"github.com/anfern777/go-serverless-framework-cdk/constants"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	cognito "github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type UserPoolProps struct {
	EnvVars constants.EnvironmentVars
}

type Cognito struct {
	Construct constructs.Construct
	CognitoData
}

type CognitoData struct {
	UserPool        cognito.UserPool
	UserPoolClients []cognito.IUserPoolClient
}

func NewCognito(scope constructs.Construct, id *string, props *UserPoolProps) *Cognito {
	this := constructs.NewConstruct(scope, id)

	userPool := cognito.NewUserPool(this, jsii.String("mainpool"), &cognito.UserPoolProps{
		SelfSignUpEnabled: jsii.Bool(false),
		AutoVerify: &cognito.AutoVerifiedAttrs{
			Email: jsii.Bool(true),
		},
		SignInAliases: &cognito.SignInAliases{
			Email: jsii.Bool(true),
		},
	})

	// groups
	cognito.NewUserPoolGroup(this, jsii.String("admin-cognito-group"), &cognito.UserPoolGroupProps{
		UserPool:    userPool,
		GroupName:   jsii.String(constants.COGNITO_ADMIN_GROUP_NAME),
		Precedence:  jsii.Number(0),
		Description: jsii.String("Admin Users"),
	})
	cognito.NewUserPoolGroup(this, jsii.String("users-cognito-group"), &cognito.UserPoolGroupProps{
		UserPool:    userPool,
		GroupName:   jsii.String(constants.COGNITO_USER_GROUP_NAME),
		Precedence:  jsii.Number(1),
		Description: jsii.String("Regular users with limited access"),
	})

	// scopes
	adminsScope := cognito.NewResourceServerScope(&cognito.ResourceServerScopeProps{
		ScopeDescription: jsii.String("full access"),
		ScopeName:        jsii.String(constants.COGNITO_ADMIN_GROUP_NAME),
	})
	usersScope := cognito.NewResourceServerScope(&cognito.ResourceServerScopeProps{
		ScopeDescription: jsii.String("read access to specific API resources"),
		ScopeName:        jsii.String(constants.COGNITO_USER_GROUP_NAME),
	})

	userServer := userPool.AddResourceServer(jsii.String("resource-server"), &cognito.UserPoolResourceServerOptions{
		Identifier: jsii.String(constants.COGNITO_RESOURCE_SERVER_IDENTIFIER),
		Scopes: &[]cognito.ResourceServerScope{
			adminsScope, usersScope,
		},
	})

	userPoolClients := []cognito.IUserPoolClient{
		cognito.NewUserPoolClient(this, jsii.String("admin-client"), &cognito.UserPoolClientProps{
			UserPool: userPool,
			OAuth: &cognito.OAuthSettings{
				CallbackUrls: jsii.Strings(fmt.Sprintf("%s%s", "https://", props.EnvVars.AdminDomain)),
				LogoutUrls:   jsii.Strings(fmt.Sprintf("%s%s", "https://", props.EnvVars.AdminDomain)),
				Scopes: &[]cognito.OAuthScope{
					cognito.OAuthScope_ResourceServer(userServer, adminsScope),
					cognito.OAuthScope_OPENID(),
				},
			},
			GenerateSecret: jsii.Bool(false),
			AuthFlows: &cognito.AuthFlow{
				UserSrp: jsii.Bool(true),
			},
		}),
		cognito.NewUserPoolClient(this, jsii.String("userportal-client"), &cognito.UserPoolClientProps{
			UserPool: userPool,
			OAuth: &cognito.OAuthSettings{
				CallbackUrls: jsii.Strings(fmt.Sprintf("%s%s", "https://", props.EnvVars.UserPortalDomain)),
				LogoutUrls:   jsii.Strings(fmt.Sprintf("%s%s", "https://", props.EnvVars.UserPortalDomain)),
				Scopes: &[]cognito.OAuthScope{
					cognito.OAuthScope_ResourceServer(userServer, usersScope),
					cognito.OAuthScope_OPENID(),
				},
			},
			GenerateSecret: jsii.Bool(false),
			AuthFlows: &cognito.AuthFlow{
				UserSrp: jsii.Bool(true),
			},
		}),
	}

	if props.EnvVars.Stage == constants.DEV {
		userPoolClients = append(userPoolClients,
			cognito.NewUserPoolClient(this, jsii.String("admin-localaccess-client"), &cognito.UserPoolClientProps{
				UserPool: userPool,
				OAuth: &cognito.OAuthSettings{
					CallbackUrls: jsii.Strings(constants.LOCAL_URL_ADMIN),
					LogoutUrls:   jsii.Strings(constants.LOCAL_URL_ADMIN),
					Scopes: &[]cognito.OAuthScope{
						cognito.OAuthScope_ResourceServer(userServer, adminsScope),
						cognito.OAuthScope_OPENID(),
					},
				},
				GenerateSecret: jsii.Bool(false),
				AuthFlows: &cognito.AuthFlow{
					UserSrp:           jsii.Bool(true),
					UserPassword:      jsii.Bool(true),
					AdminUserPassword: jsii.Bool(true),
				},
			}),
			cognito.NewUserPoolClient(this, jsii.String("userportal-localaccess-client"), &cognito.UserPoolClientProps{
				UserPool: userPool,
				OAuth: &cognito.OAuthSettings{
					CallbackUrls: jsii.Strings(constants.LOCAL_URL_USERPORTAL),
					LogoutUrls:   jsii.Strings(constants.LOCAL_URL_USERPORTAL),
					Scopes: &[]cognito.OAuthScope{
						cognito.OAuthScope_ResourceServer(userServer, usersScope),
						cognito.OAuthScope_OPENID(),
					},
				},
				GenerateSecret: jsii.Bool(false),
				AuthFlows: &cognito.AuthFlow{
					UserSrp:           jsii.Bool(true),
					AdminUserPassword: jsii.Bool(true),
				},
			}))
	}

	cognito.NewUserPoolDomain(this, jsii.String("pool-domain"), &cognito.UserPoolDomainProps{
		UserPool: userPool,
		CognitoDomain: &cognito.CognitoDomainOptions{
			DomainPrefix: jsii.String(props.EnvVars.AppName),
		},
	})

	cognito.NewCfnManagedLoginBranding(this, jsii.String("admi-login-branding"), &cognito.CfnManagedLoginBrandingProps{
		UserPoolId:               userPool.UserPoolId(),
		ClientId:                 userPoolClients[0].UserPoolClientId(),
		UseCognitoProvidedValues: jsii.Bool(true),
	})
	cognito.NewCfnManagedLoginBranding(this, jsii.String("userportal-login-branding"), &cognito.CfnManagedLoginBrandingProps{
		UserPoolId:               userPool.UserPoolId(),
		ClientId:                 userPoolClients[1].UserPoolClientId(),
		UseCognitoProvidedValues: jsii.Bool(true),
	})

	for i, upc := range userPoolClients {
		cdk.NewCfnOutput(this, jsii.String(fmt.Sprintf("cognito-client-id-%d", i)), &cdk.CfnOutputProps{
			Value:      upc.UserPoolClientId(),
			ExportName: jsii.String(fmt.Sprintf("userpool-client-id-%d", i)),
		})

		cdk.NewCfnOutput(this, jsii.String(fmt.Sprintf("cognito-userpool-id-%d", i)), &cdk.CfnOutputProps{
			Value:      userPool.UserPoolId(),
			ExportName: jsii.String(fmt.Sprintf("userpool-userpool-id-%d", i)),
		})
	}

	return &Cognito{
		Construct: this,
		CognitoData: CognitoData{
			UserPool:        userPool,
			UserPoolClients: userPoolClients,
		},
	}
}
