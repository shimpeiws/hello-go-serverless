package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, req events.S3Event) error {
	return nil
}

func main() {
	lambda.Start(handler)
}
