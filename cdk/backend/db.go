package backend

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	cdk "github.com/aws/aws-cdk-go/awscdk/v2"
	dyndb "github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DbProps struct {
}

type Db struct {
	Construct constructs.Construct
	Table     dyndb.TableV2
}

func NewDb(scope constructs.Construct, id *string, props *DbProps) *Db {
	this := constructs.NewConstruct(scope, id)

	mainTable := dyndb.NewTableV2(this, jsii.String("mainTable"), &dyndb.TablePropsV2{
		PartitionKey: &dyndb.Attribute{
			Name: jsii.String("PK"),
			Type: dyndb.AttributeType_STRING,
		},
		SortKey: &dyndb.Attribute{
			Name: jsii.String("SK"),
			Type: dyndb.AttributeType_STRING,
		},
		GlobalSecondaryIndexes: &[]*dyndb.GlobalSecondaryIndexPropsV2{
			{
				IndexName: jsii.String(constants.GSI_NAME),
				PartitionKey: &dyndb.Attribute{
					Name: jsii.String("GSI_PK"),
					Type: dyndb.AttributeType_STRING,
				},
				SortKey: &dyndb.Attribute{
					Name: jsii.String("CreatedAt"),
					Type: dyndb.AttributeType(*jsii.String(string(dyndb.AttributeType_STRING))),
				},
			},
			{
				IndexName: jsii.String(constants.GSI_NAME_2),
				PartitionKey: &dyndb.Attribute{
					Name: jsii.String("CognitoID"),
					Type: dyndb.AttributeType_STRING,
				},
			},
		},
		TableClass: dyndb.TableClass_STANDARD,
		Billing: dyndb.Billing_Provisioned(&dyndb.ThroughputProps{
			ReadCapacity: dyndb.Capacity_Fixed(jsii.Number(10)),
			WriteCapacity: dyndb.Capacity_Autoscaled(&dyndb.AutoscaledCapacityOptions{
				MaxCapacity: jsii.Number(2),
			}),
		}),
	})

	cdk.NewCfnOutput(this, jsii.String("maintable"), &cdk.CfnOutputProps{
		Value: jsii.String(*mainTable.TableArn()),
	})

	return &Db{
		Construct: this,
		Table:     mainTable,
	}
}
