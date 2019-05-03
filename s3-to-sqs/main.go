package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pkg/errors"
)

type SQSBody struct {
	BucketName string `json:"BucketName"`
	Key        string `json:"Key"`
}

func handler(ctx context.Context, req events.S3Event) error {
	sqsBody := &SQSBody{BucketName: req.Records[0].S3.Bucket.Name, Key: req.Records[0].S3.Object.Key}
	jsonString, err := json.Marshal(sqsBody)
	if err != nil {
		return errors.Wrap(err, "failed to build JSON string")
	}
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	svc := sqs.New(sess)
	qURL := os.Getenv("QUEUE_URL")
	params := &sqs.SendMessageInput{
		MessageBody:  aws.String(string(jsonString)),
		QueueUrl:     &qURL,
		DelaySeconds: aws.Int64(1),
	}
	sqsRes, err := svc.SendMessage(params)
	if err != nil {
		return err
	}
	log.Print("SetMD5OfMessageBody = " + *sqsRes.MD5OfMessageBody)
	return nil
}

func main() {
	lambda.Start(handler)
}
