package main

import (
	"bytes"
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
	"github.com/minodisk/go-fix-orientation/processor"
	"github.com/pkg/errors"
	"github.com/shimpeiws/hello-go-serverless/cloudvision"
)

func createSession() *session.Session {
	return session.Must(session.NewSession(&aws.Config{
		S3ForcePathStyle: aws.Bool(true),
		Region:           aws.String(os.Getenv("REGION")),
	}))
}

func upload(file *bytes.Reader, key string) (*s3manager.UploadOutput, error) {

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

func download(bucket string, key string) (f *bytes.Reader, err error) {
	sess := createSession()

	tempFile, _ := ioutil.TempFile("/tmp", "tempfile.jpeg")
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

	rotated, err := rotateImageFile(tempFile)
	if err != nil {
		return nil, errors.Wrap(err, "file rotate error")
	}

	return rotated, err
}

func rotateImageFile(inputFile *os.File) (outputFile *bytes.Reader, err error) {
	buff := new(bytes.Buffer)
	img, err := processor.Process(inputFile)
	if err != nil {
		return nil, errors.Wrap(err, "file rotate error")
	}
	jpeg.Encode(buff, img, nil)
	reader := bytes.NewReader(buff.Bytes())

	return reader, err
}

func minInSlice(a []int32) int32 {
	min := a[0]
	for _, i := range a {
		if i < min {
			min = i
		}
	}
	return min
}

func maxInSlice(a []int32) int32 {
	max := a[0]
	for _, i := range a {
		if i > max {
			max = i
		}
	}
	return max
}

func mosaicProcessing(inputFile *bytes.Reader, minX int32, minY int32, maxX int32, maxY int32) (outputFile *bytes.Reader, err error) {
	img, err := jpeg.Decode(inputFile)
	if err != nil {
		return nil, errors.Wrap(err, "file rotate error")
	}
	bounds := img.Bounds()
	dest := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y = y + 1 {
		for x := bounds.Min.X; x < bounds.Max.X; x = x + 1 {
			dest.Set(x, y, img.At(x, y))
		}
	}

	var block int32
	block = 25
	for y := minY + (block-1)/2; y < maxY; y = y + block {
		for x := minX + (block-1)/2; x < maxX; x = x + block {
			var cr, cg, cb float32
			var alpha uint8
			for j := y - (block-1)/2; j <= y+(block-1)/2; j++ {
				for i := x - (block-1)/2; i <= x+(block-1)/2; i++ {
					if i >= 0 && j >= 0 && i < maxX && j < maxY {
						c := color.RGBAModel.Convert(img.At(int(i), int(j)))
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
					if i >= 0 && j >= 0 && i < maxX && j < maxY {
						dest.Set(int(i), int(j), color.RGBA{uint8(cr), uint8(cg), uint8(cb), alpha})
					}
				}
			}
		}
	}
	log.Print("mosaic")

	buff := new(bytes.Buffer)
	err = jpeg.Encode(buff, dest, nil)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to create buffer")
	}
	reader := bytes.NewReader(buff.Bytes())

	return reader, nil
}

func detectFace(file *bytes.Reader) (x []int32, y []int32, err error) {
	vercityX := []int32{}
	vercityY := []int32{}
	faceAnnotations, err := cloudvision.DetectFaces(file)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Failed to detect faces")
	}
	if len(faceAnnotations) == 0 {
		log.Print("No face found")
	} else {
		log.Print("Faces: ")
		log.Print(len(faceAnnotations))
		for i, annotation := range faceAnnotations {
			boundingPoly := annotation.BoundingPoly
			log.Printf("Face %d", i)
			for _, verticy := range boundingPoly.Vertices {
				vercityX = append(vercityX, verticy.X)
				vercityY = append(vercityY, verticy.Y)
				log.Print("X: ")
				log.Print(verticy.X)
				log.Print("Y: ")
				log.Print(verticy.Y)
			}
		}
	}
	return vercityX, vercityY, err
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

	vercityX, vercityY, err := detectFace(file)

	fileProcess, err := download(bucketName, key)
	if err != nil {
		return errors.Wrap(err, "Error failed to s3 download")
	}
	log.Print("fileProcess downloaded")

	minX := minInSlice(vercityX)
	minY := minInSlice(vercityY)
	maxX := maxInSlice(vercityX)
	maxY := maxInSlice(vercityY)
	uploadFile, err := mosaicProcessing(fileProcess, minX, minY, maxX, maxY)
	if err != nil {
		return errors.Wrap(err, "mosaic processing failed")
	}

	_, err = upload(uploadFile, key)
	if err != nil {
		return errors.Wrap(err, "Error failed to s3 upload")
	}
	log.Print("uploaded")

	return nil
}

func main() {
	lambda.Start(handler)
}
