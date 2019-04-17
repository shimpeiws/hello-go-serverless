package main

import (
	"context"
	"io/ioutil"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

func createSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(os.Getenv("REGION")),
	}))
}

func upload(file *os.File, key string) (*s3manager.UploadOutput, error) {
	sess := createSession()

	uploader := s3manager.NewUploader(sess)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(os.Getenv("TARGET_S3")),
		Key:    aws.String(key),
		Body:   file,
	})

	if err != nil {
		return nil, errors.Wrap(err, "failed to upload file")
	}

	return result, err
}

func download(bucket string, key string) (f *os.File, err error) {
	sess := createSession()

	tempFile, _ := ioutil.TempFile("/tmp", "tempfile")
	defer os.Remove(tempFile.Name())

	downloader := s3manager.NewDownloader(sess)

	_, err = downloader.Download(
		tempFile,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		return nil, errors.Wrap(err, "file download error")
	}

	return tempFile, err
}

func handler(ctx context.Context, req events.S3Event) error {
	log.Print("Handler Executed!!!")

	bucketName := req.Records[0].S3.Bucket.Name
	key := req.Records[0].S3.Object.Key

	log.Print("bucketName = " + bucketName)
	log.Print("key = " + key)

	file, err := download(bucketName, key)
	if err != nil {
		return errors.Wrap(err, "Error failed to s3 download")
	}
	log.Print("downloaded")

	_, err = upload(file, key)
	if err != nil {
		return errors.Wrap(err, "Error failed to s3 upload")
	}
	log.Print("uploaded")

	return nil
}

func main() {
	lambda.Start(handler)
}
