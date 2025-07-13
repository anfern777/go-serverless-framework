package backend

import (
	"github.com/anfern777/go-serverless-framework-cdk/constants"
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatch"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudwatchactions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type CloudWatchDashboardProps struct {
	EnvVars     constants.EnvironmentVars
	LambdaNames map[string]*string
}

type CloudWatchDashboard struct {
	Construct constructs.Construct
}

func NewCloudWatchDashboard(scope constructs.Construct, id *string, props *CloudWatchDashboardProps) *CloudWatchDashboard {
	this := constructs.NewConstruct(scope, id)

	// create new Cloudwatch alarm for API gateway error reponses, i.e., statuscode 500
	api5xxAlarm := awscloudwatch.NewAlarm(this, jsii.String("ApiGateway5xxAlarm"), &awscloudwatch.AlarmProps{
		Metric: awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
			Namespace:  jsii.String("AWS/ApiGateway"),
			MetricName: jsii.String("5XXError"),
			DimensionsMap: &map[string]*string{
				"ApiName": jsii.String(props.EnvVars.AppName),
			},
			Statistic: jsii.String("Sum"),
			Period:    awscdk.Duration_Minutes(jsii.Number(5)),
		}),
		EvaluationPeriods:  jsii.Number(1),
		Threshold:          jsii.Number(1),
		ComparisonOperator: awscloudwatch.ComparisonOperator_GREATER_THAN_OR_EQUAL_TO_THRESHOLD,
		AlarmDescription:   jsii.String("Alarm if API Gateway 5XX errors are detected"),
	})

	alarmTopic := awssns.NewTopic(this, jsii.String("ApiGateway5xxAlarmTopic"), &awssns.TopicProps{
		DisplayName: jsii.String("API Gateway 5XX Alarm Topic"),
	})

	alarmTopic.AddSubscription(
		awssnssubscriptions.NewEmailSubscription(jsii.String(props.EnvVars.OfficeEmail), nil),
	)

	api5xxAlarm.AddAlarmAction(awscloudwatchactions.NewSnsAction(alarmTopic))

	dashboard := awscloudwatch.NewDashboard(this, jsii.String("dashboard"), &awscloudwatch.DashboardProps{
		DashboardName: jsii.String(props.EnvVars.AppName + "-dashboard"),
	})

	for k, v := range props.LambdaNames {
		logGroupName := jsii.String("/aws/lambda/" + *v)
		dashboard.AddWidgets(
			awscloudwatch.NewLogQueryWidget(&awscloudwatch.LogQueryWidgetProps{
				LogGroupNames: &[]*string{logGroupName},
				QueryString:   jsii.String("fields @timestamp, @message | sort @timestamp desc | limit 20"),
				Title:         jsii.String("Lambda Logs: " + k),
				Width:         jsii.Number(12),
				Height:        jsii.Number(6),
			},
			),
		)
	}

	dashboard.AddWidgets(
		awscloudwatch.NewGraphWidget(&awscloudwatch.GraphWidgetProps{
			Title: jsii.String("Estimated Current Charges (USD)"),
			Left: &[]awscloudwatch.IMetric{
				awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
					Namespace:  jsii.String("AWS/Billing"),
					MetricName: jsii.String("EstimatedCharges"),
					DimensionsMap: &map[string]*string{
						"Currency": jsii.String("USD"),
					},
					Statistic: jsii.String("Maximum"),
					Period:    awscdk.Duration_Hours(jsii.Number(6)),
				}),
			},
			Width:  jsii.Number(12),
			Height: jsii.Number(6),
		}),
	)

	dashboard.AddWidgets(
		awscloudwatch.NewSingleValueWidget(&awscloudwatch.SingleValueWidgetProps{
			Title: jsii.String("Latest Estimated Total Charge (USD)"),
			Metrics: &[]awscloudwatch.IMetric{
				awscloudwatch.NewMetric(&awscloudwatch.MetricProps{
					Namespace:  jsii.String("AWS/Billing"),
					MetricName: jsii.String("EstimatedCharges"),
					DimensionsMap: &map[string]*string{
						"Currency": jsii.String("USD"),
					},
					Statistic: jsii.String("Maximum"),
					Period:    awscdk.Duration_Hours(jsii.Number(6)),
				}),
			},
			Width:  jsii.Number(6),
			Height: jsii.Number(3),
		}),
	)

	return &CloudWatchDashboard{
		Construct: this,
	}

}
