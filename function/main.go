package function

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/hsmtkk/ubiquitous-spoon/function/storage"
	"github.com/hsmtkk/ubiquitous-spoon/function/thumbnail"
)

func init() {
	functions.CloudEvent("EntryPoint", entryPoint)
}

// StorageObjectData contains metadata of the Cloud Storage object.
type StorageObjectData struct {
	Bucket         string    `json:"bucket,omitempty"`
	Name           string    `json:"name,omitempty"`
	Metageneration int64     `json:"metageneration,string,omitempty"`
	TimeCreated    time.Time `json:"timeCreated,omitempty"`
	Updated        time.Time `json:"updated,omitempty"`
}

func entryPoint(ctx context.Context, e event.Event) error {
	log.Printf("Event ID: %s", e.ID())
	log.Printf("Event Type: %s", e.Type())

	var data StorageObjectData
	if err := e.DataAs(&data); err != nil {
		return fmt.Errorf("event.DataAs: %v", err)
	}

	log.Printf("Bucket: %s", data.Bucket)
	log.Printf("File: %s", data.Name)
	log.Printf("Metageneration: %d", data.Metageneration)
	log.Printf("Created: %s", data.TimeCreated)
	log.Printf("Updated: %s", data.Updated)

	origImage := "/tmp/original"
	thumbnail := "/tmp/thumbnail"
	thumbnailSizeStr := os.Getenv("THUMBNAIL_SIZE")
	thumbnailSize, err := strconv.Atoi(thumbnailSizeStr)
	if err != nil {
		return fmt.Errorf("failed to parse %s as int; %w", thumbnailSizeStr, err)
	}
	destinationBucket := os.Getenv("DESTINATION_BUCKET")

	op, err := storage.NewOperator(ctx)
	if err != nil {
		return err
	}
	if err := download(op, data.Bucket, data.Name, origImage); err != nil {
		return err
	}
	if err := makeThumbnail(origImage, thumbnail, thumbnailSize); err != nil {
		return err
	}
	return upload(op, destinationBucket, data.Name, thumbnail)
}

func download(op storage.Operator, bucket, key string, origPath string) error {
	origImage, err := os.Create(origPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s; %w", origPath, err)
	}
	defer origImage.Close()
	return op.Download(bucket, key, origImage)
}

func upload(op storage.Operator, bucket, key string, thumbnailPath string) error {
	thumbnail, err := os.Open(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s; %w", thumbnailPath, err)
	}
	defer thumbnail.Close()
	return op.Upload(bucket, key, thumbnail)
}

func makeThumbnail(origPath, thumbnailPath string, size int) error {
	input, err := os.Open(origPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s; %w", origPath, err)
	}
	defer input.Close()

	output, err := os.Create(thumbnailPath)
	if err != nil {
		return fmt.Errorf("failed to open file %s; %w", thumbnailPath, err)
	}
	defer output.Close()

	return thumbnail.NewMaker().Make(input, output, size)
}
