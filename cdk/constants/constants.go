package constants

import (
	"fmt"
	"strings"
)

type StageName string

const (
	// existing applicatoins stages
	LOCAL StageName = "LOCAL"
	DEV   StageName = "DEV"
	PROD  StageName = "PROD"

	// context parameters
	BUILD_FRONTDEND  = "buildFrontend"
	BUILD_ADMIN      = "buildAdmin"
	BUILD_USERPORTAL = "buildUserPortal"

	// lambda envs
	LAMBDAENV_BUCKET_NAME    string = "BUCKET_NAME"
	LAMBDA_ENV_TABLE_NAME    string = "TABLE_NAME"
	LAMBDA_ENV_GSI_NAME      string = "GSI_NAME"
	LAMBDA_ENV_SQS_QUEUE_URL string = "SQS_QUEUE_URL"

	// database
	GSI_NAME   string = "MainTableGSI"
	GSI_NAME_2 string = "MainTableGSI2"

	// s3
	DOCUMENTS_BUCKET string = "documents"
	MEDIA_BUCKET     string = "media"

	// sqs
	SQS_KEY_EMAIL string = "EMAIL"

	// app info
	APP_NAME string = "go-serverless-framework"

	// local dev
	LOCAL_URL_ADMIN      string = "http://localhost:4200"
	LOCAL_URL_USERPORTAL string = "http://localhost:4201"
	LOCAL_URL_FRONTEND   string = "http://localhost:4321"

	// cognito
	COGNITO_RESOURCE_SERVER_IDENTIFIER string = "go-serverless-framework-api"
	COGNITO_ADMIN_GROUP_NAME           string = "admin"
	COGNITO_USER_GROUP_NAME            string = "user"
)

var (
	COGNITO_ADMIN_SCOPE string = COGNITO_RESOURCE_SERVER_IDENTIFIER + "/" + "admin"
	COGNITO_USER_SCOPE  string = COGNITO_RESOURCE_SERVER_IDENTIFIER + "/" + "user"
)

type EnvironmentVars struct {
	Stage StageName

	// outputs (stage-specific if needed)
	OutputUserPoolID       string
	OutputUserPoolClientID string
	MediaURL               string

	// aws account (stage-specific if needed)
	AWSRegion    string
	AWSAccountID string

	// ses (stage-specific if needed)
	ApplicationEmail string
	OfficeEmail      string
	SESDomain        string

	// domain settings (stage-specific)
	DomainName string

	// app (stage-specific)
	AppName          string
	FrontendDomain   string
	AdminDomain      string
	UserPortalDomain string
	APIDomain        string
	MediaDomain      string

	// calculated values (stage-specific)
	AdminURL       string
	UserPortalURL  string
	FrontendURL    string
	FrontendURLWWW string
	APIURL         string

	AdminSubdomain      string
	UserPortalSubdomain string
	FrontendSubdomain   string
	APISubdomain        string
	MediaSubdomain      string
}

func NewEnvironmentVars(stage StageName) EnvironmentVars {
	var ev EnvironmentVars

	switch stage {
	case LOCAL:
		ev.Stage = LOCAL
		ev.OutputUserPoolID = "test-pool"
		ev.OutputUserPoolClientID = "test-userpoolclient-id"
		ev.MediaURL = "https://media-go-serverless-framework-test.upwigo.com"
		ev.AWSRegion = "us-east-1"
		ev.AWSAccountID = "000000000000"
		ev.ApplicationEmail = "test@upwigo.com"
		ev.OfficeEmail = "test@upwigo.com"
		ev.SESDomain = "ses-go-serverless-framework-test.upwigo.com"
		ev.DomainName = "upwigo.com"
		ev.AppName = "go-serverless-framework-local"
		ev.FrontendDomain = "go-serverless-framework-test.upwigo.com"
		ev.AdminDomain = "admin-go-serverless-framework-test.upwigo.com"
		ev.UserPortalDomain = "userportal-go-serverless-framework-test.upwigo.com"
		ev.APIDomain = "api-go-serverless-framework-test.upwigo.com"
		ev.MediaDomain = "media-go-serverless-framework-test.upwigo.com"
	case DEV:
		ev.Stage = DEV
		ev.OutputUserPoolID = "eu-central-1_tbsdJnhqN"
		ev.OutputUserPoolClientID = "4gunhuv22vp0agsi5edtgc2akf"
		ev.MediaURL = "https://media-go-serverless-framework-dev.upwigo.com"
		ev.AWSRegion = "eu-central-1"
		ev.AWSAccountID = ""
		ev.ApplicationEmail = "test@email.com"
		ev.OfficeEmail = "test@upwigo.com"
		ev.SESDomain = "ses-go-serverless-framework-dev.upwigo.com"
		ev.DomainName = "upwigo.com"
		ev.AppName = "go-serverless-framework-dev"
		ev.FrontendDomain = "go-serverless-framework-dev.upwigo.com"
		ev.AdminDomain = "admin-go-serverless-framework-dev.upwigo.com"
		ev.UserPortalDomain = "userportal-go-serverless-framework-dev.upwigo.com"
		ev.APIDomain = "api-go-serverless-framework-dev.upwigo.com"
		ev.MediaDomain = "media-go-serverless-framework-dev.upwigo.com"
	case PROD:

	default:
		fmt.Println("Unknown stage:", stage)
		ev.Stage = DEV
		ev.OutputUserPoolID = ""
		ev.OutputUserPoolClientID = ""
		ev.MediaURL = "https://media-go-serverless-framework.upwigo.com"
		ev.AWSRegion = ""
		ev.AWSAccountID = ""
		ev.ApplicationEmail = "test@upwigo.com"
		ev.OfficeEmail = "test@upwigo.com"
		ev.SESDomain = "ses-go-serverless-framework-dev.upwigo.com"
		ev.DomainName = "upwigo.com"
		ev.AppName = "go-serverless-framework"
		ev.FrontendDomain = "go-serverless-framework-staging.upwigo.com"
		ev.AdminDomain = "admin-go-serverless-framework-staging.upwigo.com"
		ev.UserPortalDomain = "userportal-go-serverless-framework-staging.upwigo.com"
		ev.APIDomain = "api-go-serverless-framework.upwigo.com"
		ev.MediaDomain = "media-go-serverless-framework.upwigo.com"
	}

	ev.AdminURL = fmt.Sprintf("https://%s", ev.AdminDomain)
	ev.UserPortalURL = fmt.Sprintf("https://%s", ev.UserPortalDomain)
	ev.FrontendURL = fmt.Sprintf("https://%s", ev.FrontendDomain)
	ev.FrontendURLWWW = fmt.Sprintf("https://www.%s", ev.FrontendDomain)
	ev.APIURL = fmt.Sprintf("https://%s", ev.APIDomain)

	ev.AdminSubdomain, _ = strings.CutSuffix(ev.AdminDomain, fmt.Sprintf("%s%s", ".", ev.DomainName))
	ev.UserPortalSubdomain, _ = strings.CutSuffix(ev.AdminDomain, fmt.Sprintf("%s%s", ".", ev.DomainName))
	ev.FrontendSubdomain, _ = strings.CutSuffix(ev.FrontendDomain, fmt.Sprintf("%s%s", ".", ev.DomainName))
	ev.APISubdomain, _ = strings.CutSuffix(ev.APIDomain, fmt.Sprintf("%s%s", ".", ev.DomainName))
	ev.MediaSubdomain, _ = strings.CutSuffix(ev.MediaDomain, fmt.Sprintf("%s%s", ".", ev.DomainName))

	return ev
}
