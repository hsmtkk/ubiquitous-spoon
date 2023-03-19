package storage

import (
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
)

type Operator interface {
	Download(bucket, key string, writer io.Writer) error
	Upload(bucket, key string, reader io.Reader) error
}

func NewOperator(ctx context.Context) (Operator, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to make Cloud Storage client; %w", err)
	}
	return &operatorImpl{ctx, client}, nil
}

type operatorImpl struct {
	ctx    context.Context
	client *storage.Client
}

func (o *operatorImpl) Download(bucket, key string, writer io.Writer) error {
	bkt := o.client.Bucket(bucket)
	obj := bkt.Object(key)
	reader, err := obj.NewReader(o.ctx)
	if err != nil {
		return fmt.Errorf("failed to open object; %s; %s; %s", bucket, key, err)
	}
	defer reader.Close()
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy contents; %w", err)
	}
	return nil
}

func (o *operatorImpl) Upload(bucket, key string, reader io.Reader) error {
	bkt := o.client.Bucket(bucket)
	obj := bkt.Object(key)
	writer := obj.NewWriter(o.ctx)
	defer writer.Close()
	if _, err := io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy contents; %w", err)
	}
	return nil
}
