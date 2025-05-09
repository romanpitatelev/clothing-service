package objectsrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	transport "github.com/aws/smithy-go/endpoints"
	"github.com/romanpitatelev/clothing-service/internal/entity"
)

type S3Config struct {
	Address string
	Bucket  string
	Access  string
	Secret  string
	Region  string
}

type resolver struct {
	URL *url.URL
}

func (r *resolver) ResolveEndpoint(_ context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	u := *r.URL
	u.Path += "/" + *params.Bucket

	return transport.Endpoint{URI: u}, nil
}

type S3 struct {
	cfg    S3Config
	client *s3.Client
}

func New(cfg S3Config) (*S3, error) {
	endpoint, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse S3 address: %w", err)
	}

	client := s3.New(s3.Options{
		Credentials:        credentials.NewStaticCredentialsProvider(cfg.Access, cfg.Secret, ""),
		Region:             cfg.Region,
		EndpointResolverV2: &resolver{URL: endpoint},
	})

	s3Repo := S3{
		cfg:    cfg,
		client: client,
	}

	return &s3Repo, nil
}

func (s *S3) UploadFile(data []byte, fileName, contentType string) error {
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &s.cfg.Bucket,
		Key:         aws.String(fileName),
		Body:        bytes.NewReader(data),
		ContentType: &contentType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *S3) UploadReader(data io.ReadSeeker, fileName string) error {
	buff := make([]byte, 512)
	if _, err := data.Read(buff); err != nil {
		return fmt.Errorf("failed to read file mime type: %w", err)
	}

	mimeType := http.DetectContentType(buff)
	if _, err := data.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to seek file: %w", err)
	}

	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      &s.cfg.Bucket,
		Key:         aws.String(fileName),
		Body:        data,
		ContentType: &mimeType,
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (s *S3) DownloadFile(fileName string) (io.ReadCloser, string, error) {
	result, err := s.client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &s.cfg.Bucket,
		Key:    aws.String(fileName),
	})

	var responseError *awshttp.ResponseError

	switch {
	case err == nil:
	case errors.As(err, &responseError) && responseError.HTTPStatusCode() == http.StatusNotFound:
		return nil, "", entity.ErrFileNotFound
	default:
		return nil, "", fmt.Errorf("failed to download file: %w", err)
	}

	return result.Body, *result.ContentType, nil
}

func (s *S3) DeleteFiles(fileNames ...string) error {
	toRemove := make([]types.ObjectIdentifier, 0, len(fileNames))
	for _, fileName := range fileNames {
		toRemove = append(toRemove, types.ObjectIdentifier{Key: aws.String(fileName)})
	}

	_, err := s.client.DeleteObjects(context.Background(), &s3.DeleteObjectsInput{
		Bucket: &s.cfg.Bucket,
		Delete: &types.Delete{Objects: toRemove},
	})
	if err != nil {
		return fmt.Errorf("failed to delete files: %w", err)
	}

	return nil
}

func (s *S3) ListAllFiles(prefix string) ([]string, error) {
	result, err := s.client.ListObjectsV2(context.Background(), &s3.ListObjectsV2Input{
		Bucket: &s.cfg.Bucket,
		Prefix: aws.String(prefix),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	files := make([]string, 0, len(result.Contents))
	for _, object := range result.Contents {
		files = append(files, *object.Key)
	}

	return files, nil
}
