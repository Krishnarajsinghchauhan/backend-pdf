package queue

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

var sqsClient *sqs.Client

// Initialize SQS client
func Init() {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Println("Failed to load AWS config:", err)
		return
	}

	sqsClient = sqs.NewFromConfig(cfg)
}

// Push job to SQS queue
func PushJob(queueURL string, job interface{}) {
	if sqsClient == nil {
		log.Println("SQS client is nil! Did you forget queue.Init() ?")
		return
	}

	body, _ := json.Marshal(job)

	_, err := sqsClient.SendMessage(context.TODO(), &sqs.SendMessageInput{
		QueueUrl:    &queueURL,
		MessageBody: newString(string(body)),
	})

	if err != nil {
		log.Println("Failed to push job:", err)
	} else {
		log.Println("Job pushed to:", queueURL)
	}
}

func newString(s string) *string {
	return &s
}
