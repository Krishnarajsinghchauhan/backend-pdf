package handlers

import (
	"context"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
)

type CreateUploadURLRequest struct {
	FileName string `json:"fileName"`
}

func CreateUploadURL(c *fiber.Ctx) error {
	var body CreateUploadURLRequest

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// ✅ FORCE the correct region (us-east-1)
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-east-1"),
	)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "AWS config error"})
	}

	client := s3.NewFromConfig(cfg)
	bucket := os.Getenv("AWS_S3_BUCKET")

	// file key
	key := "uploads/" + time.Now().Format("20060102150405") + "_" + body.FileName

	presigner := s3.NewPresignClient(client)

	// generate presigned PUT url
	resp, err := presigner.PresignPutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create presigned URL"})
	}

	// return URL to frontend
	return c.JSON(fiber.Map{
		"url":      resp.URL,                          // PUT to this URL
		"file_url": "https://" + bucket + ".s3.amazonaws.com/" + key, // public url
	})
}
