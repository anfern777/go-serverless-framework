package frontend

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	r53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	r53targets "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53targets"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	s3deploy "github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type FrontendCdkStackProps struct {
	awscdk.StackProps
	HostedZone     r53.IHostedZone
	CloudFrontCert *acm.DnsValidatedCertificate
	EnvVars        constants.EnvironmentVars
}

func NewFrontendCdkStack(scope constructs.Construct, id string, props *FrontendCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	if props.EnvVars.Stage == constants.PROD {
		sprops.StackName = jsii.String(id)
	}

	stack := awscdk.NewStack(scope, &id, &sprops)

	bucket := s3.NewBucket(stack, jsii.String("hosting-bucket"), &s3.BucketProps{
		Encryption:        s3.BucketEncryption_S3_MANAGED,
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY,
		AutoDeleteObjects: jsii.Bool(true),
	})

	cf := NewCloudFront(stack, jsii.String("distribution"), &CloudFrontProps{
		HostingBucket:  bucket,
		CloudFrontCert: props.CloudFrontCert,
		EnvVars:        props.EnvVars,
	})

	target := r53targets.NewCloudFrontTarget(*cf.distribution)

	rootRecordName := jsii.String("")
	if props.EnvVars.Stage == constants.DEV {
		rootRecordName = jsii.String(props.EnvVars.FrontendDomain)
	}

	r53.NewARecord(stack, jsii.String("frontend-arecord"), &r53.ARecordProps{
		Zone:       props.HostedZone,
		Target:     r53.RecordTarget_FromAlias(target),
		RecordName: rootRecordName,
	})

	if props.EnvVars.Stage == constants.PROD {
		r53.NewCnameRecord(stack, jsii.String("www-arecord"), &r53.CnameRecordProps{
			RecordName: jsii.String("www"),
			Zone:       props.HostedZone,
			DomainName: jsii.String(*cf.CfDomainName),
		})
	}

	dirName, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	buildApp, err := strconv.ParseBool(stack.Node().GetContext(jsii.String(constants.BUILD_FRONTDEND)).(string))
	if err != nil {
		panic(err)
	}

	if buildApp {
		var deploymentEnv string
		if props.EnvVars.Stage == constants.DEV {
			deploymentEnv = "development"
		} else {
			deploymentEnv = "production"
		}
		s3deploy.NewBucketDeployment(stack, jsii.String("s3-deploy"), &s3deploy.BucketDeploymentProps{
			Sources: &[]s3deploy.ISource{
				s3deploy.Source_Asset(jsii.String(filepath.Join(dirName, "..", "frontend")), &awss3assets.AssetOptions{
					Bundling: &awscdk.BundlingOptions{
						User: jsii.String("root"),
						Image: awscdk.NewDockerImage(
							jsii.String("node:22"),
							jsii.String("sha256:0e910f435308c36ea60b4cfd7b80208044d77a074d16b768a81901ce938a62dc"),
						),
						Command: &[]*string{
							jsii.String("sh"),
							jsii.String("-c"),
							jsii.String(fmt.Sprintf(`
								npm i && \
								npm run build -- --mode %s && \
								cp -r dist/* /asset-output/`, deploymentEnv)),
						},
						WorkingDirectory: jsii.String("/asset-input"),
						OutputType:       awscdk.BundlingOutput_NOT_ARCHIVED,
					},
				}),
			},
			DestinationBucket: bucket,
		})
	}

	awscdk.Tags_Of(cf.Construct).Add(jsii.String("Module"), jsii.String("app-hosting"), nil)

	return stack
}
