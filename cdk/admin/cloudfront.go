package admin

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	acm "github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	cf "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfront"
	origins "github.com/aws/aws-cdk-go/awscdk/v2/awscloudfrontorigins"
	s3 "github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CloudFrontProps struct {
	HostingBucket  s3.IBucket
	CloudFrontCert *acm.DnsValidatedCertificate
	EnvVars        constants.EnvironmentVars
}

type CloudFront struct {
	constructs.Construct
	CfDomainName *string
}

func NewCloudFront(scope constructs.Construct, id *string, props *CloudFrontProps) *CloudFront {
	this := constructs.NewConstruct(scope, id)

	oai := cf.NewOriginAccessIdentity(this, jsii.String("oai"), &cf.OriginAccessIdentityProps{})

	props.HostingBucket.GrantRead(oai, nil)

	distribution := cf.NewDistribution(this, jsii.String("distribution"), &cf.DistributionProps{
		DomainNames: &[]*string{
			jsii.String(props.EnvVars.AdminDomain),
		},
		Certificate: *props.CloudFrontCert,
		DefaultBehavior: &cf.BehaviorOptions{
			Origin: origins.S3BucketOrigin_WithOriginAccessIdentity(props.HostingBucket, &origins.S3BucketOriginWithOAIProps{
				OriginAccessIdentity: oai,
			}),
		},
		ErrorResponses: &[]*cf.ErrorResponse{
			{
				HttpStatus:         jsii.Number(403),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
			},
			{
				HttpStatus:         jsii.Number(404),
				ResponseHttpStatus: jsii.Number(200),
				ResponsePagePath:   jsii.String("/index.html"),
			},
		},
		DefaultRootObject: jsii.String("index.html"),
	})

	awscdk.NewCfnOutput(this, jsii.String("cf-domain"), &awscdk.CfnOutputProps{
		Value:      distribution.DomainName(),
		ExportName: jsii.String("admin-domain"),
	})

	return &CloudFront{
		this,
		distribution.DomainName(),
	}
}
