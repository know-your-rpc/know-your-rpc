package s3

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func ListS3Objects(bucket string) ([]string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}

	var keys []string
	paginator := s3.NewListObjectsV2Paginator(client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(context.TODO())
		if err != nil {
			return nil, fmt.Errorf("failed to get page: %w", err)
		}

		for _, object := range page.Contents {
			keys = append(keys, *object.Key)
		}
	}

	return keys, nil
}

func WriteS3Object(bucket string, key string, data []byte) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   bytes.NewReader(data),
	}

	_, err = client.PutObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to write object: %w", err)
	}

	return nil
}

// TODO: add caching
func ReadS3Object(bucket string, key string) ([]byte, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("error loading AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := client.GetObject(context.TODO(), input)
	if err != nil {
		return nil, fmt.Errorf("failed to get object: %w", err)
	}
	defer result.Body.Close()

	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object body: %w", err)
	}

	return content, nil
}

// TODO: don't initialize s3 over and over again
func CopyS3Object(bucket string, key string, newBucket string, newKey string) error {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return fmt.Errorf("error loading AWS configuration: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	source := bucket + "/" + key

	input := &s3.CopyObjectInput{
		Bucket:     aws.String(newBucket),
		CopySource: aws.String(source),
		Key:        aws.String(newKey),
	}

	_, err = client.CopyObject(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to copy object, %w", err)
	}

	return nil
}
