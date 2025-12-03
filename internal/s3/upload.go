package s3

import (
    "bytes"
    "io"
    "mime"
    "path/filepath"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/s3"
)

var bucket = "pdf-master-storage" // ← your bucket name

// UploadFile uploads a file to S3 and returns the public URL
func UploadFile(file io.Reader, key string) (string, error) {
    sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-1"),
    })
    if err != nil {
        return "", err
    }

    svc := s3.New(sess)

    buf := new(bytes.Buffer)
    if _, err := io.Copy(buf, file); err != nil {
        return "", err
    }

    // Detect content type from file extension
    contentType := mime.TypeByExtension(filepath.Ext(key))
    if contentType == "" {
        contentType = "application/octet-stream"
    }

    _, err = svc.PutObject(&s3.PutObjectInput{
        Bucket:      aws.String(bucket),
        Key:         aws.String(key),
        Body:        bytes.NewReader(buf.Bytes()),
        ContentType: aws.String(contentType),
        ACL:         aws.String("public-read"), // public preview
    })
    if err != nil {
        return "", err
    }

    return "https://" + bucket + ".s3.amazonaws.com/" + key, nil
}
