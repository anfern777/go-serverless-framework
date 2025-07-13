package backend

import (
	"fmt"
	"net/http"

	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	r53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	r53targets "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type BackendCdkStackProps struct {
	awscdk.StackProps
	EnvVars constants.EnvironmentVars
}

type BackendStack struct {
	awscdk.Stack
	HostedZone       r53.IHostedZone
	UserPoolId       *string
	UserPoolClientId *string
	CloudFrontCert   *acm.DnsValidatedCertificate
}

func NewBackendCdkStack(scope constructs.Construct, id string, props *BackendCdkStackProps) *BackendStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	if props.EnvVars.Stage == constants.PROD {
		sprops.StackName = jsii.String(id)
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	cognito := NewCognito(stack, jsii.String("user-pool"), &UserPoolProps{
		EnvVars: props.EnvVars,
	})

	buckets := make(map[string]awss3.IBucket)
	buckets[constants.DOCUMENTS_BUCKET] = awss3.NewBucket(stack, jsii.String(constants.DOCUMENTS_BUCKET), &awss3.BucketProps{
		Encryption:    awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
		Cors: &[]*awss3.CorsRule{
			{
				AllowedHeaders: jsii.Strings("*"),
				ExposedHeaders: jsii.Strings("Etag"),
				AllowedMethods: &[]awss3.HttpMethods{
					http.MethodGet,
					http.MethodHead,
					http.MethodPut,
				},
				AllowedOrigins: jsii.Strings(
					constants.LOCAL_URL_ADMIN,
					constants.LOCAL_URL_FRONTEND,
					props.EnvVars.AdminURL,
					props.EnvVars.FrontendURL,
					props.EnvVars.FrontendURLWWW,
				),
			},
		},
	})

	buckets[constants.MEDIA_BUCKET] = awss3.NewBucket(stack, jsii.String(constants.MEDIA_BUCKET), &awss3.BucketProps{
		Encryption:    awss3.BucketEncryption_S3_MANAGED,
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
		Cors: &[]*awss3.CorsRule{
			{
				AllowedHeaders: jsii.Strings("*"),
				ExposedHeaders: jsii.Strings("Etag"),
				AllowedMethods: &[]awss3.HttpMethods{
					http.MethodGet,
					http.MethodHead,
					http.MethodPut,
				},
				AllowedOrigins: jsii.Strings(
					constants.LOCAL_URL_ADMIN,
					constants.LOCAL_URL_FRONTEND,
					props.EnvVars.AdminURL,
					props.EnvVars.FrontendURL,
					props.EnvVars.FrontendURLWWW,
				),
			},
		},
	},
	)

	hz := r53.HostedZone_FromLookup(stack, jsii.String("hz"), &r53.HostedZoneProviderProps{
		DomainName: jsii.String(props.EnvVars.DomainName),
	})

	cloudFrontCert := acm.NewDnsValidatedCertificate(stack, jsii.String("public-cert-cf"), &acm.DnsValidatedCertificateProps{
		DomainName: jsii.String(props.EnvVars.DomainName),
		Region:     jsii.String("us-east-1"),
		SubjectAlternativeNames: &[]*string{
			jsii.String(fmt.Sprintf("*.%s", props.EnvVars.DomainName)),
		},
		HostedZone: hz,
		Validation: acm.CertificateValidation_FromDns(hz),
	})

	mediaCf := NewCloudFront(stack, jsii.String("cf-media"), &CloudFrontProps{
		HostingBucket:  buckets[constants.MEDIA_BUCKET],
		CloudFrontCert: &cloudFrontCert,
		EnvVars:        props.EnvVars,
	})

	target := r53targets.NewCloudFrontTarget(*mediaCf.distribution)

	r53.NewARecord(stack, jsii.String("media-arecord"), &r53.ARecordProps{
		Zone:       hz,
		Target:     r53.RecordTarget_FromAlias(target),
		RecordName: jsii.String(props.EnvVars.MediaSubdomain),
	})

	certAPIGateway := acm.NewCertificate(stack, jsii.String("public-cert-apigateway"), &acm.CertificateProps{
		DomainName: jsii.String(props.EnvVars.DomainName),
		SubjectAlternativeNames: &[]*string{
			jsii.String(fmt.Sprintf("*.%s", props.EnvVars.DomainName)),
		},
		Validation: acm.CertificateValidation_FromDns(hz),
	})

	db := NewDb(stack, jsii.String("database"), nil)

	ses := NewSes(stack, jsii.String("email-service"), &SesProps{
		EnvVars: props.EnvVars,
	})

	sqs := NewSQS(stack, jsii.String("sqs"), &SQSProps{
		EnvVars: props.EnvVars,
	})

	api := NewApi(stack, jsii.String("api"), &ApiProps{
		Buckets:         buckets,
		DbTable:         db.Table,
		SesArn:          ses.Arn,
		CognitoArn:      cognito.UserPool.UserPoolArn(),
		Certificate:     certAPIGateway,
		HostedZone:      hz,
		UserPool:        cognito.UserPool,
		UserPoolClients: cognito.UserPoolClients,
		EnvVars:         props.EnvVars,
		SQSQueues:       sqs.Queues,
	})

	NewCloudWatchDashboard(stack, jsii.String("dashboard"), &CloudWatchDashboardProps{
		EnvVars:     props.EnvVars,
		LambdaNames: api.LambdaNames,
	})

	for i, b := range buckets {
		awscdk.NewCfnOutput(
			stack,
			jsii.String(fmt.Sprintf("%s-%s", i, "-bucket-name")),
			&awscdk.CfnOutputProps{
				Value: b.BucketName(),
			})
	}

	awscdk.NewCfnOutput(stack, jsii.String("CloudFrontMediaUrl"), &awscdk.CfnOutputProps{
		Value: mediaCf.CfDomainName,
	})
	awscdk.NewCfnOutput(stack, jsii.String("HostedZoneId"), &awscdk.CfnOutputProps{
		Value: hz.HostedZoneId(),
	})

	awscdk.Tags_Of(mediaCf.Construct).Add(jsii.String("Module"), jsii.String("storage"), nil)
	awscdk.Tags_Of(cognito.Construct).Add(jsii.String("Module"), jsii.String("user-management"), nil)
	awscdk.Tags_Of(db.Construct).Add(jsii.String("Module"), jsii.String("database"), nil)
	awscdk.Tags_Of(api.Construct).Add(jsii.String("Module"), jsii.String("api"), nil)
	awscdk.Tags_Of(ses.Construct).Add(jsii.String("Module"), jsii.String("ses"), nil)

	return &BackendStack{
		Stack:            stack,
		HostedZone:       hz,
		UserPoolId:       cognito.UserPool.UserPoolId(),
		UserPoolClientId: cognito.UserPool.UserPoolId(),
		CloudFrontCert:   &cloudFrontCert,
	}
}
