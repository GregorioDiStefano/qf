package clouds

import (
	"io"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type AWSStorage struct {
	bucket string
	s3     *s3.S3
}

func NewAWSStorage(bucket string) *AWSStorage {
	creds := credentials.NewEnvCredentials()
	sess, err := session.NewSession(&aws.Config{Credentials: creds})

	if err != nil {
		panic(err)
	}

	// Create S3 service client
	svc := s3.New(sess)
	return &AWSStorage{bucket, svc}
}

func (awsS3 *AWSStorage) ListObjects() ([]string, error) {
	input := &s3.ListObjectsInput{}
	objects := []string{}

	result, err := awsS3.s3.ListObjects(input)

	for _, obj := range result.Contents {
		objects = append(objects, obj.GoString())
	}

	return objects, err
}

func (awsS3 *AWSStorage) Remove(obj string) error {
	_, err := awsS3.s3.DeleteObject(&s3.DeleteObjectInput{Key: aws.String(obj), Bucket: aws.String(awsS3.bucket)})
	return err
}

func (awsS3 *AWSStorage) Download(obj string) (io.Reader, error) {
	input := &s3.GetObjectInput{Key: aws.String(obj), Bucket: aws.String(awsS3.bucket)}
	result, err := awsS3.s3.GetObject(input)
	return result.Body, err
}

func (awsS3 *AWSStorage) Upload(obj string, r io.Reader) error {
	uploader := s3manager.NewUploaderWithClient(awsS3.s3)

	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(awsS3.bucket),
		Key:    aws.String(obj),
		Body:   r,
	})

	return err
}
