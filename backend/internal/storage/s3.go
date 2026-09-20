package storage

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Store struct {
	bucket string
	client *s3.Client
}

func NewS3Store(ctx context.Context, region string, bucket string) (*S3Store, error) {
	if region == "" {
		return nil, errors.New("AWS_REGION is required")
	}
	if bucket == "" {
		return nil, errors.New("AWS_S3_BUCKET is required")
	}

	// The AWS SDK reads credentials from environment variables, AWS profiles, or cloud roles.
	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(region))
	if err != nil {
		return nil, err
	}

	return &S3Store{
		bucket: bucket,
		client: s3.NewFromConfig(cfg),
	}, nil
}

func (s *S3Store) PutObject(ctx context.Context, input PutObjectInput) (ObjectMetadata, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(input.Key),
		Body:          input.Body,
		ContentType:   aws.String(input.ContentType),
		ContentLength: aws.Int64(input.Size),
	})
	if err != nil {
		return ObjectMetadata{}, err
	}

	return ObjectMetadata{
		Bucket:      s.bucket,
		Key:         input.Key,
		ContentType: input.ContentType,
		Size:        input.Size,
	}, nil
}

func (s *S3Store) GetObject(ctx context.Context, key string) (GetObjectOutput, error) {
	output, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return GetObjectOutput{}, err
	}

	return GetObjectOutput{
		Body:        output.Body,
		ContentType: aws.ToString(output.ContentType),
		Size:        aws.ToInt64(output.ContentLength),
	}, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// Why this file exists:
// This is the S3 implementation of ObjectStore.
// S3 stores binary files as objects inside a bucket; each object is found by its key.
// io.Reader lets Go stream file bytes without loading the whole file into memory first.
