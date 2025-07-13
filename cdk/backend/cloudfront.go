package backend

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
	distribution *cf.Distribution
}

func NewCloudFront(scope constructs.Construct, id *string, props *CloudFrontProps) *CloudFront {
	this := constructs.NewConstruct(scope, id)

	oai := cf.NewOriginAccessIdentity(this, jsii.String("oai"), &cf.OriginAccessIdentityProps{})

	props.HostingBucket.GrantRead(oai, nil)

	distribution := cf.NewDistribution(this, jsii.String("media-distribution"), &cf.DistributionProps{
		DomainNames: &[]*string{
			jsii.String(props.EnvVars.MediaDomain),
		},
		Certificate: *props.CloudFrontCert,
		DefaultBehavior: &cf.BehaviorOptions{
			Origin: origins.S3BucketOrigin_WithOriginAccessIdentity(props.HostingBucket, &origins.S3BucketOriginWithOAIProps{
				OriginAccessIdentity: oai,
			}),
			ViewerProtocolPolicy: cf.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
		},
	})

	awscdk.NewCfnOutput(this, jsii.String("media-cf-domain"), &awscdk.CfnOutputProps{
		Value:      distribution.DomainName(),
		ExportName: jsii.String("media-cf-domain"),
	})

	return &CloudFront{
		this,
		distribution.DomainName(),
		&distribution,
	}
}
