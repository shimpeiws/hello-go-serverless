package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/pkg/errors"
	"github.com/shimpeiws/hello-go-serverless/s3"
)

type SQSBody struct {
	UserID string `json:"UserID"`
	Image  string `json:"Image"`
}

func handler(ctx context.Context, req events.SQSEvent) error {
	log.Print("Handler Executed!!!")
	for _, record := range req.Records {
		log.Print("queue body = " + record.Body)
		var sqsBody SQSBody
		recordBodyBytes := []byte(record.Body)
		if err := json.Unmarshal(recordBodyBytes, &sqsBody); err != nil {
			return errors.Wrap(err, "failed to parse SQS Body")
		}
		data, err := base64.StdEncoding.DecodeString(sqsBody.Image)
		if err != nil {
			return errors.Wrap(err, "failed to decode image")
		}
		key := sqsBody.UserID + ".jpg"
		reader := bytes.NewReader(data)
		log.Print("key = " + key)
		res, err := s3.Upload(reader, os.Getenv("TARGET_S3"), key)
		if err != nil {
			return errors.Wrap(err, "failed to upload image")
		}
		log.Print("res = " + res.UploadID)
	}
	return nil
}

func main() {
	lambda.Start(handler)
}
