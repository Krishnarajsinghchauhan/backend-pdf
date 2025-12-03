package s3uploader

import (
    "bytes"
    "context"
    "mime"
    "os"
    "path/filepath"

    "github.com/aws/aws-sdk-go-v2/aws"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func UploadPublicFile(buffer []byte, key string) (string, error) {
    bucket := os.Getenv("AWS_S3_BUCKET")

    cfg, err := config.LoadDefaultConfig(context.TODO(),
        config.WithRegion("us-east-1"),
    )
    if err != nil {
        return "", err
    }

    client := s3.NewFromConfig(cfg)

    contentType := mime.TypeByExtension(filepath.Ext(key))
    if contentType == "" {
        contentType = "image/png"
    }

    _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(key),
        Body:        bytes.NewReader(buffer),
        ContentType: aws.String(contentType),
        ACL:         aws.String("public-read"),
    })

    if err != nil {
        return "", err
    }

    return "https://" + bucket + ".s3.amazonaws.com/" + key, nil
}
