package middleware

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
)

type Handler func(ctx context.Context, req events.APIGatewayV2HTTPRequest) (*events.APIGatewayV2HTTPResponse, error)
