package clouds

import (
	"context"
	"encoding/base64"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"path/filepath"

	"cloud.google.com/go/storage"
	// Imports the Google Cloud Storage client package.
)

const (
	GOOGLE_CREDS_FILENAME = "creds.json"
)

type GoogleStorage struct {
	bucket string
	client *storage.Client
}

func setupGoogle(googleCreds string) {
	//make sure cred file is there
	data, err := base64.StdEncoding.DecodeString(googleCreds)

	if err != nil {
		panic("failed to read base64'ed google credentials: " + err.Error())
	}

	usr, err := user.Current()

	// abort if we can get a user's home directory
	if err != nil {
		log.Fatal(err)
	}

	credsLocations := filepath.Join(usr.HomeDir, GOOGLE_CREDS_FILENAME)

	if err := ioutil.WriteFile(credsLocations, []byte(data), 0600); err != nil {
		panic("failed to write google credential file: " + err.Error())
	}

	os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsLocations)
}

func NewGoogleStorage(bucket, googleCreds string) *GoogleStorage {
	ctx := context.Background()
	setupGoogle(googleCreds)

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

func (gs *GoogleStorage) ListObjectsWithPrefix(prefix string) ([]string, error) {
	ctx := context.Background()
	objs := []string{}

	q := storage.Query{Prefix: prefix}
	next := gs.client.Bucket(gs.bucket).Objects(ctx, &q)

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
