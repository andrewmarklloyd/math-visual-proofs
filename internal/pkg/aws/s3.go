package aws

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Client struct {
	s3 *s3.Client
}

func NewClient() (Client, error) {
	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(os.Getenv("AWS_ACCESS_KEY_ID"), os.Getenv("AWS_SECRET_ACCESS_KEY"), "")),
	)
	if err != nil {
		return Client{}, fmt.Errorf("loading default config: %s", err)
	}

	cfg.Region = os.Getenv("AWS_REGION")
	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:           os.Getenv("SPACES_URL"),
			SigningRegion: os.Getenv("AWS_REGION"),
		}, nil
	})

	client := s3.NewFromConfig(cfg)

	return Client{
		s3: client,
	}, nil
}

func (c *Client) UploadFile(ctx context.Context, localVideoPath, key string) error {
	file, err := os.Open(localVideoPath)
	if err != nil {
		return fmt.Errorf("opening local video file: %s", err)
	}

	uploader := manager.NewUploader(c.s3)
	_, err = uploader.Upload(ctx, &s3.PutObjectInput{
		Bucket: aws.String("math-visual-proofs"),
		Key:    aws.String(fmt.Sprintf("renderings/%s", key)),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("uploading video file to s3: %s", err)
	}

	return nil
}
