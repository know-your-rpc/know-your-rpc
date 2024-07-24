package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"

	"koonopek/know_your_rpc/writer/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const USERS_BUCKET = "know-your-rpc-users"

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

func ReadRpcUrlsForUser(userAddress string, chainId string) ([]utils.RpcInfo, error) {
	bucketKey := fmt.Sprintf("%s.json", userAddress)

	data, err := ReadS3Object(USERS_BUCKET, bucketKey)

	if err != nil {
		var noSuchKeyError *types.NoSuchKey
		if errors.As(err, &noSuchKeyError) {
			fmt.Printf("copying public.json to %s, because it was not created yet", bucketKey)
			err := CopyS3Object(USERS_BUCKET, "public.json", USERS_BUCKET, bucketKey)
			if err != nil {
				return nil, fmt.Errorf("failed to copy s3 object error=%s", err)
			}
			data, err = ReadS3Object(USERS_BUCKET, bucketKey)
			if err != nil {
				return nil, fmt.Errorf("failed to read from s3 after copying it bucket=%s bucketKey=%s err=%s", USERS_BUCKET, bucketKey, err)
			}
		} else {
			return nil, fmt.Errorf("failed to read from s3 bucket=%s bucketKey=%s err=%s", USERS_BUCKET, bucketKey, err)
		}
	}

	rpcInfoMap := &utils.RpcInfoMap{}
	err = json.Unmarshal(data, rpcInfoMap)
	if err != nil {
		return nil, fmt.Errorf("failed to parse chain list %v", err)
	}

	rpcUrls, ok := (*rpcInfoMap)[chainId]

	if !ok {
		return nil, fmt.Errorf("couldn't find rpcUrls for chainId=%s", chainId)
	}

	return rpcUrls, nil
}
