package main

import (
	"github.com/anfern777/go-serverless-framework-cdk/admin"
	"github.com/anfern777/go-serverless-framework-cdk/backend"
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/anfern777/go-serverless-framework-cdk/constants/frontend"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// Define the stage
type AppStageProps struct {
	awscdk.StageProps
	envVars constants.EnvironmentVars
}

func NewDevStage(scope constructs.Construct, id string, props *AppStageProps) awscdk.Stage {
	stage := awscdk.NewStage(scope, &id, &props.StageProps)

	backendStack := backend.NewBackendCdkStack(stage, "go-serverless-framework-backend", &backend.BackendCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars: props.envVars,
	})

	frontend.NewFrontendCdkStack(stage, "go-serverless-framework-frontend", &frontend.FrontendCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars:        props.envVars,
		HostedZone:     backendStack.HostedZone,
		CloudFrontCert: backendStack.CloudFrontCert,
	})

	admin.NewAdminCdkStack(stage, "go-serverless-framework-admin", &admin.AdminCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars:        props.envVars,
		CloudFrontCert: backendStack.CloudFrontCert,
		HostedZone:     backendStack.HostedZone,
	})

	awscdk.Tags_Of(backendStack.Stack).Add(jsii.String("App"), jsii.String("go-serverless-framework"), nil)
	awscdk.Tags_Of(backendStack.Stack).Add(jsii.String("Environment"), jsii.String("dev"), nil)

	return stage
}

func NewProdStage(scope constructs.Construct, id string, props *AppStageProps) awscdk.Stage {
	stage := awscdk.NewStage(scope, &id, &props.StageProps)

	backendStack := backend.NewBackendCdkStack(stage, "go-serverless-framework-backend", &backend.BackendCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars: props.envVars,
	})

	frontend.NewFrontendCdkStack(stage, "go-serverless-framework-frontend", &frontend.FrontendCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars:        props.envVars,
		HostedZone:     backendStack.HostedZone,
		CloudFrontCert: backendStack.CloudFrontCert,
	})

	admin.NewAdminCdkStack(stage, "go-serverless-framework-admin", &admin.AdminCdkStackProps{
		StackProps: awscdk.StackProps{
			Env: env(props.envVars),
		},
		EnvVars:        props.envVars,
		CloudFrontCert: backendStack.CloudFrontCert,
		HostedZone:     backendStack.HostedZone,
	})

	awscdk.Tags_Of(backendStack.Stack).Add(jsii.String("App"), jsii.String("go-serverless-framework"), nil)
	awscdk.Tags_Of(backendStack.Stack).Add(jsii.String("Environment"), jsii.String("production"), nil)

	return stage
}
