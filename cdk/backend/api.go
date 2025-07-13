package backend

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/anfern777/go-serverless-framework-cdk/constants"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	apigateway "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	authorizers "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2authorizers"
	v2integrations "github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2integrations"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	cognito "github.com/aws/aws-cdk-go/awscdk/v2/awscognito"
	dyndb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	iam "github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambdaeventsources"
	r53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	r53targets "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	lambda "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ApiProps struct {
	Buckets         map[string]s3.IBucket
	SQSQueues       map[string]awssqs.Queue
	DbTable         dyndb.TableV2
	SesArn          *string
	CognitoArn      *string
	Certificate     acm.Certificate
	HostedZone      r53.IHostedZone
	UserPool        cognito.IUserPool
	UserPoolClients []cognito.IUserPoolClient
	EnvVars         constants.EnvironmentVars
}

type Api struct {
	Construct   constructs.Construct
	ApiUrl      *string
	LambdaNames map[string]*string
}

func NewApi(scope constructs.Construct, id *string, props *ApiProps) *Api {
	this := constructs.NewConstruct(scope, id)

	dirName, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// ------------ LAMBDA FUNCTIONS ------------

	lambdaFunctionsConfig := map[string]map[string]map[string]any{
		"applications": {
			"getBeteewnDates": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "get_between_dates"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME),
				},
			},
			"getDocumentContent": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "get_document_content"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"getSingle": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "get_single"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
				},
			},
			"getDocuments": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "get_documents"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
				},
			},
			"postDocumentsRequest": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "post_documents_request"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
				},
			},
			"putDocumentUpload": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "put_document_upload"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
				},
			},
			"getMessages": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "get_messages"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
				},
			},
			"postMessage": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "post_message"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME_2),
				},
			},
			"create": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "create"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME):    jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME):    jsii.String(*props.DbTable.TableName()),
					*jsii.String("OFFICE_EMAIL"):                     jsii.String(props.EnvVars.OfficeEmail),
					*jsii.String("APPLICATION_EMAIL"):                jsii.String(props.EnvVars.ApplicationEmail),
					*jsii.String(constants.LAMBDA_ENV_SQS_QUEUE_URL): jsii.String(*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueUrl()),
				},
			},
			"delete": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "delete"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"deleteDocument": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "delete_document"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.DOCUMENTS_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"putProperty": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "put_property"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"putDocumentProperty": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "applications", "put_document_property"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
		},
		"posts": {
			"getByLang": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "posts", "get_by_lang"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
					*jsii.String(constants.LAMBDA_ENV_GSI_NAME):   jsii.String(constants.GSI_NAME),
				},
			},
			"getSingle": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "posts", "get_single"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"create": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "posts", "create"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.MEDIA_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"update": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "posts", "update"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.MEDIA_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"delete": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "posts", "delete"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME): jsii.String(*props.Buckets[constants.MEDIA_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
		},
		"cognito": {
			"createUser": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "cognito", "create_user"),
				"env": &map[string]*string{
					*jsii.String("USERPOOL_ID"):                   jsii.String(*props.UserPool.UserPoolId()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
			"deleteUser": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "cognito", "delete_user"),
				"env": &map[string]*string{
					*jsii.String("USERPOOL_ID"):                   jsii.String(*props.UserPool.UserPoolId()),
					*jsii.String(constants.LAMBDA_ENV_TABLE_NAME): jsii.String(*props.DbTable.TableName()),
				},
			},
		},
		"contact": {
			"sendEmail": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "contact"),
				"env": &map[string]*string{
					*jsii.String("OFFICE_EMAIL"):                     jsii.String(props.EnvVars.OfficeEmail),
					*jsii.String("APPLICATION_EMAIL"):                jsii.String(props.EnvVars.ApplicationEmail),
					*jsii.String(constants.LAMBDA_ENV_SQS_QUEUE_URL): jsii.String(*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueUrl()),
				},
			},
		},
		"sendEmployerDeclaration": {
			"sendEmail": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "routes", "send_employer_declaration"),
				"env": &map[string]*string{
					*jsii.String("APPLICATION_EMAIL"):                jsii.String(props.EnvVars.ApplicationEmail),
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME):    jsii.String(*props.Buckets[constants.MEDIA_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_SQS_QUEUE_URL): jsii.String(*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueUrl()),
				},
			},
		},
		"eventHandlers": {
			"sqsEmail": {
				"path": filepath.Join(dirName, "..", "api", "cmd", "event_handlers", "sqs", "email"),
				"env": &map[string]*string{
					*jsii.String(constants.LAMBDAENV_BUCKET_NAME):    jsii.String(*props.Buckets[constants.MEDIA_BUCKET].BucketName()),
					*jsii.String(constants.LAMBDA_ENV_SQS_QUEUE_URL): jsii.String(*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueUrl()),
				},
			},
		},
	}

	lambdaFunctions := make(map[string]lambda.GoFunction)
	for tableName, lambdaConfig := range lambdaFunctionsConfig {
		for key, data := range lambdaConfig {
			uniqueKey := fmt.Sprintf("%s-%s", tableName, key)
			lambdaFunctions[uniqueKey] = lambda.NewGoFunction(this, jsii.String(fmt.Sprintf("%s-%s", tableName, key)), &lambda.GoFunctionProps{
				Entry:       jsii.String(data["path"].(string)),
				Environment: data["env"].(*map[string]*string),
				// TODO: add non deprecated version of log retention
				// LogRetention: awslogs.RetentionDays(fmt.Sprint(30)),
				Tracing:    awslambda.Tracing_DISABLED,
				Timeout:    cdk.Duration_Seconds(jsii.Number(3)),
				MemorySize: jsii.Number(128),
			})
		}
	}

	lambdaNames := make(map[string]*string)
	for key, lambdaFunction := range lambdaFunctions {
		lambdaNames[key] = lambdaFunction.FunctionName()
	}

	// ------------ EVENT SOURCES ------------
	sqqEmailEventSource := awslambdaeventsources.NewSqsEventSource(props.SQSQueues[constants.SQS_KEY_EMAIL], &awslambdaeventsources.SqsEventSourceProps{
		BatchSize:         jsii.Number(2),
		MaxBatchingWindow: cdk.Duration_Seconds(jsii.Number(20)),
	})
	lambdaFunctions["eventHandlers-sqsEmail"].AddEventSource(sqqEmailEventSource)

	// ------------ IAM PERMISSIONS DEFNITION ------------

	// S3
	s3PermissionsConfig := map[string]map[string]string{
		constants.DOCUMENTS_BUCKET: {
			"get":    "s3:GetObject",
			"put":    "s3:PutObject",
			"delete": "s3:DeleteObject",
		},
		constants.MEDIA_BUCKET: {
			"get":    "s3:GetObject",
			"put":    "s3:PutObject",
			"delete": "s3:DeleteObject",
		},
	}
	s3Permissions := make(map[string]iam.PolicyStatement)
	for tableName, permissions := range s3PermissionsConfig {
		for permissionKey, permission := range permissions {
			uniqueKey := fmt.Sprintf("%s-%s", tableName, permissionKey)
			s3Permissions[uniqueKey] = iam.NewPolicyStatement(&iam.PolicyStatementProps{
				Resources: jsii.Strings(
					fmt.Sprintf("%s/*", *props.Buckets[tableName].BucketArn()),
				),
				Actions: jsii.Strings(permission),
			})
		}
	}
	lambdaFunctions["applications-getDocumentContent"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.DOCUMENTS_BUCKET, "get")])
	lambdaFunctions["applications-putDocumentUpload"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.DOCUMENTS_BUCKET, "put")])
	lambdaFunctions["applications-create"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.DOCUMENTS_BUCKET, "put")])
	lambdaFunctions["applications-delete"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.DOCUMENTS_BUCKET, "delete")])
	lambdaFunctions["applications-deleteDocument"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.DOCUMENTS_BUCKET, "delete")])
	lambdaFunctions["posts-create"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.MEDIA_BUCKET, "put")])
	lambdaFunctions["posts-update"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.MEDIA_BUCKET, "put")])
	lambdaFunctions["posts-delete"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.MEDIA_BUCKET, "delete")])
	lambdaFunctions["eventHandlers-sqsEmail"].AddToRolePolicy(s3Permissions[fmt.Sprintf("%s-%s", constants.MEDIA_BUCKET, "get")])

	// DATABASE
	dbPermissionsConfig := map[string]map[string]string{
		"get": {
			"action": "dynamodb:GetItem",
			"table":  *props.DbTable.TableArn(),
		},
		"put": {
			"action": "dynamodb:PutItem",
			"table":  *props.DbTable.TableArn(),
		},
		"update": {
			"action": "dynamodb:UpdateItem",
			"table":  *props.DbTable.TableArn(),
		},
		"delete": {
			"action": "dynamodb:DeleteItem",
			"table":  *props.DbTable.TableArn(),
		},
		"query-mt": {
			"action": "dynamodb:Query",
			"table":  *props.DbTable.TableArn(),
		},
		"query-gsi": {
			"action": "dynamodb:Query",
			"table":  fmt.Sprintf("%s/index/%s", *props.DbTable.TableArn(), constants.GSI_NAME),
		},
		"query-gsi2": {
			"action": "dynamodb:Query",
			"table":  fmt.Sprintf("%s/index/%s", *props.DbTable.TableArn(), constants.GSI_NAME_2),
		},
		"batchWriteItem": {
			"action": "dynamodb:BatchWriteItem",
			"table":  *props.DbTable.TableArn(),
		},
		"conditionalCheckItem": {
			"action": "dynamodb:ConditionCheckItem",
			"table":  *props.DbTable.TableArn(),
		},
	}

	dbPermissions := make(map[string]iam.PolicyStatement)
	for permissionKey, permissionData := range dbPermissionsConfig {
		dbPermissions[permissionKey] = iam.NewPolicyStatement(&iam.PolicyStatementProps{
			Resources: jsii.Strings(
				permissionData["table"],
			),
			Actions: jsii.Strings(permissionData["action"]),
		})
	}

	lambdaFunctions["applications-getDocumentContent"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["applications-putDocumentUpload"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["applications-putDocumentUpload"].AddToRolePolicy(dbPermissions["put"])
	lambdaFunctions["applications-putDocumentUpload"].AddToRolePolicy(dbPermissions["query-gsi2"])
	lambdaFunctions["applications-putDocumentUpload"].AddToRolePolicy(dbPermissions["conditionalCheckItem"])
	lambdaFunctions["applications-create"].AddToRolePolicy(dbPermissions["put"])
	lambdaFunctions["applications-create"].AddToRolePolicy(dbPermissions["batchWriteItem"])
	lambdaFunctions["applications-postDocumentsRequest"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["applications-postDocumentsRequest"].AddToRolePolicy(dbPermissions["batchWriteItem"])
	lambdaFunctions["applications-getBeteewnDates"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["applications-getBeteewnDates"].AddToRolePolicy(dbPermissions["query-gsi"])
	lambdaFunctions["applications-delete"].AddToRolePolicy(dbPermissions["delete"])
	lambdaFunctions["applications-deleteDocument"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["applications-deleteDocument"].AddToRolePolicy(dbPermissions["delete"])
	lambdaFunctions["applications-delete"].AddToRolePolicy(dbPermissions["batchWriteItem"])
	lambdaFunctions["applications-delete"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["applications-delete"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["applications-putProperty"].AddToRolePolicy(dbPermissions["update"])
	lambdaFunctions["applications-putDocumentProperty"].AddToRolePolicy(dbPermissions["update"])
	lambdaFunctions["applications-getDocuments"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["applications-getDocuments"].AddToRolePolicy(dbPermissions["query-gsi2"])
	lambdaFunctions["applications-getMessages"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["applications-getMessages"].AddToRolePolicy(dbPermissions["query-gsi2"])
	lambdaFunctions["applications-postMessage"].AddToRolePolicy(dbPermissions["put"])
	lambdaFunctions["applications-postMessage"].AddToRolePolicy(dbPermissions["query-gsi2"])
	lambdaFunctions["applications-postMessage"].AddToRolePolicy(dbPermissions["conditionalCheckItem"])

	lambdaFunctions["posts-getSingle"].AddToRolePolicy(dbPermissions["get"])
	lambdaFunctions["posts-getByLang"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["posts-getByLang"].AddToRolePolicy(dbPermissions["query-gsi"])
	lambdaFunctions["posts-create"].AddToRolePolicy(dbPermissions["put"])
	lambdaFunctions["posts-create"].AddToRolePolicy(dbPermissions["batchWriteItem"])
	lambdaFunctions["posts-update"].AddToRolePolicy(dbPermissions["update"])
	lambdaFunctions["posts-delete"].AddToRolePolicy(dbPermissions["delete"])
	lambdaFunctions["posts-delete"].AddToRolePolicy(dbPermissions["query-mt"])
	lambdaFunctions["posts-delete"].AddToRolePolicy(dbPermissions["batchWriteItem"])

	lambdaFunctions["cognito-createUser"].AddToRolePolicy(dbPermissions["update"])
	lambdaFunctions["cognito-deleteUser"].AddToRolePolicy(dbPermissions["update"])

	// SES
	sesSendEmailPermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
		Resources: jsii.Strings(
			*props.SesArn,
		),
		Actions: jsii.Strings("ses:SendRawEmail"),
	})
	lambdaFunctions["eventHandlers-sqsEmail"].AddToRolePolicy(sesSendEmailPermissions)

	// COGNITO
	cognitoCreatePermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
		Resources: jsii.Strings(
			*props.CognitoArn,
		),
		Actions: jsii.Strings("cognito-idp:AdminCreateUser"),
	})
	cognitoAddUserToGroupPermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
		Resources: jsii.Strings(
			*props.CognitoArn,
		),
		Actions: jsii.Strings("cognito-idp:AdminAddUserToGroup"),
	})
	cognitoDeletePermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
		Resources: jsii.Strings(
			*props.CognitoArn,
		),
		Actions: jsii.Strings("cognito-idp:AdminDeleteUser"),
	})
	lambdaFunctions["cognito-createUser"].AddToRolePolicy(cognitoCreatePermissions)
	lambdaFunctions["cognito-createUser"].AddToRolePolicy(cognitoAddUserToGroupPermissions)
	lambdaFunctions["cognito-deleteUser"].AddToRolePolicy(cognitoDeletePermissions)

	// ------------------- SQS --------------------
	sqsSendPermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
		Resources: jsii.Strings(
			*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueArn(),
		),
		Actions: jsii.Strings("sqs:SendMessage"),
	})
	// sqsBatchSendPermissions := iam.NewPolicyStatement(&iam.PolicyStatementProps{
	// 	Resources: jsii.Strings(
	// 		*props.SQSQueues[constants.SQS_KEY_EMAIL].QueueArn(),
	// 	),
	// 	Actions: jsii.Strings("sqs:SendMessageBatch"),
	// })
	lambdaFunctions["contact-sendEmail"].AddToRolePolicy(sqsSendPermissions)
	lambdaFunctions["applications-postMessage"].AddToRolePolicy(sqsSendPermissions)
	lambdaFunctions["applications-postDocumentsRequest"].AddToRolePolicy(sqsSendPermissions)
	lambdaFunctions["sendEmployerDeclaration-sendEmail"].AddToRolePolicy(sqsSendPermissions)
	lambdaFunctions["applications-create"].AddToRolePolicy(sqsSendPermissions)

	// ------------ API GATEWAY CONFIG ------------
	customDomain := apigateway.NewDomainName(this, jsii.String("api-domain"), &apigateway.DomainNameProps{
		DomainName:  jsii.String(props.EnvVars.APIDomain),
		Certificate: props.Certificate,
	})

	corsOrigins := []string{props.EnvVars.AdminURL, props.EnvVars.FrontendURL, props.EnvVars.FrontendURLWWW}

	if props.EnvVars.Stage == constants.DEV {
		corsOrigins = append(corsOrigins, constants.LOCAL_URL_ADMIN, constants.LOCAL_URL_FRONTEND)
	}

	httpApi := apigateway.NewHttpApi(this, jsii.String("HttpApi"), &apigateway.HttpApiProps{
		ApiName:            jsii.String("go-serverless-framework-api"),
		CreateDefaultStage: jsii.Bool(true),
		DefaultDomainMapping: &apigateway.DomainMappingOptions{
			DomainName: customDomain,
		},
		CorsPreflight: &apigateway.CorsPreflightOptions{
			AllowMethods: &[]apigateway.CorsHttpMethod{
				apigateway.CorsHttpMethod_ANY,
			},
			AllowOrigins: jsii.Strings(corsOrigins...),
			AllowHeaders: jsii.Strings("Content-Type", "Authorization", "*"),
		},
	})

	authorizer := authorizers.NewHttpUserPoolAuthorizer(jsii.String("userpool-authorizer"), props.UserPool, &authorizers.HttpUserPoolAuthorizerProps{
		UserPoolClients: &props.UserPoolClients,
	})

	alias := r53targets.NewApiGatewayv2DomainProperties(customDomain.RegionalDomainName(), customDomain.RegionalHostedZoneId())
	r53.NewARecord(this, jsii.String("api-a-record"), &r53.ARecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String(props.EnvVars.APISubdomain),
		Target:     r53.RecordTarget_FromAlias(alias),
	})

	// STAGES
	defaultStage := httpApi.DefaultStage()
	defaultStageResource := defaultStage.Node().DefaultChild()
	if cfnStage, ok := defaultStageResource.(apigateway.CfnStage); ok {
		cfnStage.AddPropertyOverride(jsii.String("DefaultRouteSettings.ThrottlingBurstLimit"), 2)
		cfnStage.AddPropertyOverride(jsii.String("DefaultRouteSettings.ThrottlingRateLimit"), 1)
	} else {
		panic("Failed to cast to CfnStage")
	}

	// ROUTES
	routesConfig := map[string][]map[string]any{
		"applications": {
			{
				"route":       "/gt-application",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_create-Applications"), lambdaFunctions["applications-create"], nil),
				"authorizer":  false,
			},
			{
				"route":       "/gt-applications/date-range/{startDate}/{endDate}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_get-between-dates-Applications"), lambdaFunctions["applications-getBeteewnDates"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/{id}",
				"method":      apigateway.HttpMethod_DELETE,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_delete-Applications"), lambdaFunctions["applications-delete"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/document/content/{id}/{document-type}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getDocumentContent-Applications"), lambdaFunctions["applications-getDocumentContent"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/document/upload/{id}",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("putDocumentUpload-Applications"), lambdaFunctions["applications-putDocumentUpload"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/document/upload",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("putDocumentUpload-Applications"), lambdaFunctions["applications-putDocumentUpload"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE, constants.COGNITO_USER_SCOPE},
			},
			{
				"route":       "/gt-application/property/{id}",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_updateProperty-Applications"), lambdaFunctions["applications-putProperty"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/document/property/{id}/{document-type}",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_putDocumentProperty-Applications"), lambdaFunctions["applications-putDocumentProperty"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/document/{id}/{document-type}",
				"method":      apigateway.HttpMethod_DELETE,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_deleteDocument-Applications"), lambdaFunctions["applications-deleteDocument"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/documents/{id}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getDocuments-Applications"), lambdaFunctions["applications-getDocuments"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/documents",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getDocuments-Applications"), lambdaFunctions["applications-getDocuments"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE, constants.COGNITO_USER_SCOPE},
			},
			{
				"route":       "/gt-application/documents/request/{id}",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("postDocumentsRequest-Applications"), lambdaFunctions["applications-postDocumentsRequest"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/messages/{id}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getMessages-Applications"), lambdaFunctions["applications-getMessages"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/messages",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getMessages-Applications"), lambdaFunctions["applications-getMessages"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE, constants.COGNITO_USER_SCOPE},
			},
			{
				"route":       "/gt-application/{id}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getSingle-Applications"), lambdaFunctions["applications-getSingle"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getSingle-Applications"), lambdaFunctions["applications-getSingle"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE, constants.COGNITO_USER_SCOPE},
			},
			{
				"route":       "/gt-application/message/{id}",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_postMessage-Applications"), lambdaFunctions["applications-postMessage"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/gt-application/message",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getMessages-Applications"), lambdaFunctions["applications-postMessage"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE, constants.COGNITO_USER_SCOPE},
			},
		},
		"posts": {
			{
				"route":       "/post/language/{language}",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_create-posts"), lambdaFunctions["posts-create"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/post/{id}",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_update-posts"), lambdaFunctions["posts-update"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/post/{id}",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_getSingle-posts"), lambdaFunctions["posts-getSingle"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
			{
				"route":       "/posts",
				"method":      apigateway.HttpMethod_GET,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_get-posts"), lambdaFunctions["posts-getByLang"], nil),
				"authorizer":  false,
			},
			{
				"route":       "/post/{id}",
				"method":      apigateway.HttpMethod_DELETE,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_delete-posts"), lambdaFunctions["posts-delete"], nil),
				"authorizer":  true,
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
			},
		},
		"cognito": {
			{
				"route":       "/cognito/create",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_cognito-createUser"), lambdaFunctions["cognito-createUser"], nil),
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
				"authorizer":  true,
			},
			{
				"route":       "/cognito/delete",
				"method":      apigateway.HttpMethod_PUT,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("_cognito-deleteUser"), lambdaFunctions["cognito-deleteUser"], nil),
				"scopes":      []string{*cognito.OAuthScope_COGNITO_ADMIN().ScopeName(), constants.COGNITO_ADMIN_SCOPE},
				"authorizer":  true,
			},
		},
		"contact": {
			{
				"route":       "/contact",
				"method":      apigateway.HttpMethod_POST,
				"integration": v2integrations.NewHttpLambdaIntegration(jsii.String("contact-sendEmail"), lambdaFunctions["contact-sendEmail"], nil),
				"authorizer":  false,
			},
		},
	}

	for _, routeConfig := range routesConfig {
		for _, param := range routeConfig {
			options := &apigateway.AddRoutesOptions{
				Path: jsii.String(param["route"].(string)),
				Methods: &[]apigateway.HttpMethod{
					param["method"].(apigateway.HttpMethod),
				},
				Integration: param["integration"].(v2integrations.HttpLambdaIntegration),
			}

			// Check if an authorizer is specified for this route
			if hasAuth, exists := param["authorizer"]; exists && hasAuth.(bool) {
				options.Authorizer = authorizer
				options.AuthorizationScopes = jsii.Strings(param["scopes"].([]string)...)
			}
			httpApi.AddRoutes(options)
		}
	}

	// OUTPUTS
	cdk.NewCfnOutput(this, jsii.String("api-endpoint"), &cdk.CfnOutputProps{
		Value:      jsii.String(*httpApi.Url()),
		ExportName: jsii.String("api-endpoint"),
	})

	return &Api{
		Construct:   this,
		ApiUrl:      httpApi.Url(),
		LambdaNames: lambdaNames,
	}
}
