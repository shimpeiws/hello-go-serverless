package main

import (
	"context"
	"image"
	"image/color"
	"image/jpeg"
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

	"github.com/shimpeiws/hello-go-serverless/cloudvision"
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

	faceAnnotations, err := cloudvision.DetectFaces(ctx, file)
	if err != nil {
		return errors.Wrap(err, "Failed to detect faces")
	}
	if len(faceAnnotations) == 0 {
		log.Print("No face found")
	} else {
		log.Print("Faces: ")
		for i, annotation := range faceAnnotations {
			boundingPoly := annotation.BoundingPoly
			log.Printf("Face %d", i)
			for _, verticy := range boundingPoly.Vertices {
				log.Printf("X: %d", verticy.X)
				log.Printf("Y: %d", verticy.Y)
			}
		}
	}

	img, _, err := image.Decode(file)
	if err != nil {
		return errors.Wrap(err, "decode error")
	}
	log.Print("decoded")
	bounds := img.Bounds()
	log.Print("bounds")
	dest := image.NewRGBA(bounds)
	log.Print("dest")
	block := 11
	for y := bounds.Min.Y + (block-1)/2; y < bounds.Max.Y; y = y + block {
		for x := bounds.Min.X + (block-1)/2; x < bounds.Max.X; x = x + block {
			var cr, cg, cb float32
			var alpha uint8
			for j := y - (block-1)/2; j <= y+(block-1)/2; j++ {
				for i := x - (block-1)/2; i <= x+(block-1)/2; i++ {
					if i >= 0 && j >= 0 && i < bounds.Max.X && j < bounds.Max.Y {
						c := color.RGBAModel.Convert(img.At(i, j))
						col := c.(color.RGBA)
						cr += float32(col.R)
						cg += float32(col.G)
						cb += float32(col.B)
						alpha = col.A
					}
				}
			}
			cr = cr / float32(block*block)
			cg = cg / float32(block*block)
			cb = cb / float32(block*block)
			for j := y - (block-1)/2; j <= y+(block-1)/2; j++ {
				for i := x - (block-1)/2; i <= x+(block-1)/2; i++ {
					if i >= 0 && j >= 0 && i < bounds.Max.X && j < bounds.Max.Y {
						dest.Set(i, j, color.RGBA{uint8(cr), uint8(cg), uint8(cb), alpha})
					}
				}
			}
		}
	}
	tempFile, _ := ioutil.TempFile("/tmp", "tempOutfile")
	defer os.Remove(tempFile.Name())
	err = jpeg.Encode(tempFile, dest, nil)
	if err != nil {
		return errors.Wrap(err, "Error failed to encode image")
	}

	_, err = upload(tempFile, key)
	if err != nil {
		return errors.Wrap(err, "Error failed to s3 upload")
	}
	log.Print("uploaded")

	return nil
}

func main() {
	lambda.Start(handler)
}
