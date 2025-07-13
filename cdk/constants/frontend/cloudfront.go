package frontend

import (
	"fmt"
	"os"
	"path/filepath"

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

	dirName, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	routerHandlerFunction := cf.NewFunction(this, jsii.String("RouterHandler"), &cf.FunctionProps{
		Code: cf.FunctionCode_FromFile(&cf.FileCodeOptions{
			FilePath: jsii.String(filepath.Join(dirName, "frontend", "cf-functions", "router-handler.ts")),
		}),
		AutoPublish:  jsii.Bool(true),
		FunctionName: jsii.String("router-handler"),
		Runtime:      cf.FunctionRuntime_JS_2_0(),
	})

	wwwRedirectFunction := cf.NewFunction(this, jsii.String("wwwRedirectFunc"), &cf.FunctionProps{
		Code: cf.FunctionCode_FromFile(&cf.FileCodeOptions{
			FilePath: jsii.String(filepath.Join(dirName, "frontend", "cf-functions", "redirect-www.ts")),
		}),
		AutoPublish:  jsii.Bool(true),
		FunctionName: jsii.String("redirect-www"),
		Runtime:      cf.FunctionRuntime_JS_2_0(),
	})

	domainName := []*string{
		jsii.String(props.EnvVars.FrontendDomain),
	}

	cfFunctions := []*cf.FunctionAssociation{&cf.FunctionAssociation{
		EventType: cf.FunctionEventType_VIEWER_REQUEST,
		Function:  routerHandlerFunction,
	}}

	if props.EnvVars.Stage == constants.PROD {
		cfFunctions = append(cfFunctions, &cf.FunctionAssociation{
			EventType: cf.FunctionEventType_VIEWER_REQUEST,
			Function:  wwwRedirectFunction,
		})
		domainName = append(domainName, jsii.String(fmt.Sprintf("%s%s", "www.", props.EnvVars.FrontendDomain)))
	}

	distribution := cf.NewDistribution(this, jsii.String("distribution"), &cf.DistributionProps{
		DomainNames: &domainName,
		Certificate: *props.CloudFrontCert,
		DefaultBehavior: &cf.BehaviorOptions{
			Origin: origins.S3BucketOrigin_WithOriginAccessIdentity(props.HostingBucket, &origins.S3BucketOriginWithOAIProps{
				OriginAccessIdentity: oai,
			}),
			ViewerProtocolPolicy: cf.ViewerProtocolPolicy_REDIRECT_TO_HTTPS,
			FunctionAssociations: &cfFunctions,
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
		ExportName: jsii.String("frontend-domain"),
	})

	return &CloudFront{
		this,
		distribution.DomainName(),
		&distribution,
	}
}
