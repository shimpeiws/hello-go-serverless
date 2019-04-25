package cloudvision

import (
	"bytes"
	"context"
	"log"

	vision "cloud.google.com/go/vision/apiv1"
	visionpb "google.golang.org/genproto/googleapis/cloud/vision/v1"
)

func client(ctx context.Context) (*vision.ImageAnnotatorClient, error) {
	client, err := vision.NewImageAnnotatorClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	return client, err
}

func DetectFaces(ctx context.Context, file *bytes.Reader) ([]*visionpb.FaceAnnotation, error) {
	image, err := vision.NewImageFromReader(file)
	if err != nil {
		log.Fatalf("Failed to read image: %v", err)
	}

	client, err := client(ctx)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	annotations, err := client.DetectFaces(ctx, image, nil, 10)
	if err != nil {
		log.Fatalf("Failed to annotations: %v", err)
	}

	return annotations, err
}
