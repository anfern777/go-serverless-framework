package backend

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssqs"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type SQSProps struct {
	EnvVars constants.EnvironmentVars
}

type SQS struct {
	Construct constructs.Construct
	Queues    map[string]awssqs.Queue
}

func NewSQS(scope constructs.Construct, id *string, props *SQSProps) *SQS {
	this := constructs.NewConstruct(scope, id)

	queues := make(map[string]awssqs.Queue, 1)

	queues[constants.SQS_KEY_EMAIL] = awssqs.NewQueue(this, jsii.String("emailQueue"), &awssqs.QueueProps{
		DeadLetterQueue: &awssqs.DeadLetterQueue{
			Queue:           awssqs.NewQueue(this, jsii.String("emailDDQ"), &awssqs.QueueProps{}),
			MaxReceiveCount: jsii.Number(3),
		},
		EnforceSSL:             jsii.Bool(true),
		ReceiveMessageWaitTime: cdk.Duration_Seconds(jsii.Number(20)),
	})

	for queueKey, queue := range queues {
		cdk.NewCfnOutput(this, jsii.String(queueKey+"sqsqueue"), &cdk.CfnOutputProps{
			Value:      jsii.String(*queue.QueueArn()),
			ExportName: jsii.String(queueKey),
		})
	}

	return &SQS{
		this,
		queues,
	}
}
