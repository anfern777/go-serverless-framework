package admin

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	r53 "github.com/aws/aws-cdk-go/awscdk/v2/awsroute53"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3assets"
	s3deploy "github.com/aws/aws-cdk-go/awscdk/v2/awss3deployment"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type AdminCdkStackProps struct {
	awscdk.StackProps
	CloudFrontCert *acm.DnsValidatedCertificate
	HostedZone     r53.IHostedZone
	EnvVars        constants.EnvironmentVars
}

func NewAdminCdkStack(scope constructs.Construct, id string, props *AdminCdkStackProps) awscdk.Stack {
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

	r53.NewCnameRecord(stack, jsii.String("admin-cname"), &r53.CnameRecordProps{
		Zone:       props.HostedZone,
		RecordName: jsii.String(props.EnvVars.AdminSubdomain),
		DomainName: cf.CfDomainName,
	})

	dirName, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	buildApp, err := strconv.ParseBool(stack.Node().GetContext(jsii.String(constants.BUILD_ADMIN)).(string))
	if err != nil {
		panic(err)
	}

	config := "production"
	if props.EnvVars.Stage == constants.DEV {
		config = "staging"
	}

	if buildApp {
		s3deploy.NewBucketDeployment(stack, jsii.String("s3-deploy"), &s3deploy.BucketDeploymentProps{
			Sources: &[]s3deploy.ISource{
				s3deploy.Source_Asset(jsii.String(filepath.Join(dirName, "..", "admin")), &awss3assets.AssetOptions{
					Bundling: &awscdk.BundlingOptions{
						User:  jsii.String("root"),
						Image: awscdk.NewDockerImage(jsii.String("node:22"), jsii.String("sha256:0e910f435308c36ea60b4cfd7b80208044d77a074d16b768a81901ce938a62dc")),
						Command: &[]*string{
							jsii.String("sh"),
							jsii.String("-c"),
							jsii.String(fmt.Sprintf(`
								npm i && \ 
								npm run build -- --configuration=%s && \
								cp -r dist/admin/browser/* /asset-output/`, config)),
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
