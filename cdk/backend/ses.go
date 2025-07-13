package backend

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	ses "github.com/aws/aws-cdk-go/awscdk/v2/awsses"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type SesProps struct {
	EnvVars constants.EnvironmentVars
}

type Ses struct {
	Construct constructs.Construct
	Arn       *string
}

func NewSes(scope constructs.Construct, id *string, props *SesProps) *Ses {
	this := constructs.NewConstruct(scope, id)

	identity := ses.NewEmailIdentity(this, jsii.String("domain-identity"), &ses.EmailIdentityProps{
		Identity:       ses.Identity_Domain(jsii.String(props.EnvVars.DomainName)),
		DkimIdentity:   ses.DkimIdentity_EasyDkim(ses.EasyDkimSigningKeyLength_RSA_2048_BIT),
		MailFromDomain: jsii.String(props.EnvVars.SESDomain),
	})

	return &Ses{
		Construct: this,
		Arn:       identity.EmailIdentityArn(),
	}
}
