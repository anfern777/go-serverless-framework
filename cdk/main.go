package main

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

type AdminCdkStackProps struct {
	awscdk.StackProps
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Create the development stage
	localEnvVars := constants.NewEnvironmentVars(constants.LOCAL)
	NewDevStage(app, "Local", &AppStageProps{
		StageProps: awscdk.StageProps{
			Env: env(localEnvVars),
		},
		envVars: localEnvVars,
	})

	// Create the development stage
	stageEnvVars := constants.NewEnvironmentVars(constants.DEV)
	NewDevStage(app, "Dev", &AppStageProps{
		StageProps: awscdk.StageProps{
			Env: env(stageEnvVars),
		},
		envVars: stageEnvVars,
	})

	// Create the production stage
	prodEnvVars := constants.NewEnvironmentVars(constants.PROD)
	NewProdStage(app, "Prod", &AppStageProps{
		StageProps: awscdk.StageProps{
			Env: env(prodEnvVars),
		},
		envVars: prodEnvVars,
	})

	app.Synth(nil)
}

func env(envVars constants.EnvironmentVars) *awscdk.Environment {
	return &awscdk.Environment{
		Account: jsii.String(envVars.AWSAccountID),
		Region:  jsii.String(envVars.AWSRegion),
	}
}
