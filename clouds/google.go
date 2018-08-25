package clouds

import (
	"context"
	"io"
	"log"

	"cloud.google.com/go/storage"
	// Imports the Google Cloud Storage client package.
)

type GoogleStorage struct {
	bucket string
	client *storage.Client
}

func NewGoogleStorage(bucket string) *GoogleStorage {
	ctx := context.Background()
	client, err := storage.NewClient(ctx)

	if err != nil {
		log.Fatal(err)
	}

	return &GoogleStorage{bucket, client}
}

func (gs *GoogleStorage) ListObjects() ([]string, error) {
	ctx := context.Background()
	objs := []string{}

	next := gs.client.Bucket(gs.bucket).Objects(ctx, &storage.Query{})

	for {
		obj, err := next.Next()
		if obj != nil && err == nil {
			objs = append(objs, obj.Name)
		} else {
			break
		}
	}

	return objs, nil
}

func (gs *GoogleStorage) GetMetadata(obj string) (map[string]string, error) {
	ctx := context.Background()
	attrs, err := gs.client.Bucket(gs.bucket).Object(obj).Attrs(ctx)

	if err != nil {
		return nil, err
	}

	return attrs.Metadata, nil
}

func (gs *GoogleStorage) SetMetadata(obj string, metadata map[string]string) error {
	ctx := context.Background()
	data := storage.ObjectAttrsToUpdate{Metadata: metadata}
	_, err := gs.client.Bucket(gs.bucket).Object(obj).Update(ctx, data)
	return err
}

func (gs *GoogleStorage) Download(obj string) (io.Reader, error) {
	ctx := context.Background()
	rc, err := gs.client.Bucket(gs.bucket).Object(obj).NewReader(ctx)

	if err != nil {
		return nil, err
	}

	return rc, err
}

func (gs *GoogleStorage) Upload(obj string, r io.Reader) error {
	ctx := context.Background()
	w := gs.client.Bucket(gs.bucket).Object(obj).NewWriter(ctx)

	if _, err := io.Copy(w, r); err != nil {
		return err
	}

	return w.Close()
}

func (gs *GoogleStorage) Remove(obj string) error {
	ctx := context.Background()
	return gs.client.Bucket(gs.bucket).Object(obj).Delete(ctx)
}
