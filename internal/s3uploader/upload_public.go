package s3uploader

import (
    "bytes"
    "context"
    "fmt"
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

    // 🚫 NO ACL — bucket owner enforced blocks ACL usage
    _, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(key),
        Body:        bytes.NewReader(buffer),
        ContentType: aws.String(contentType),
    })

    if err != nil {
        fmt.Println("S3 PUT ERROR:", err)
        return "", err
    }

    // Public URL (your bucket policy must allow public read on previews/*)
    return "https://" + bucket + ".s3.amazonaws.com/" + key, nil
}
