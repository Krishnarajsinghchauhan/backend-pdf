package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"backend/internal/queue"
	"backend/internal/redis"
	"backend/internal/jobs"
	"encoding/json"
)

type CreateJobRequest struct {
	Tool     string   `json:"tool"`
	FileURLs []string `json:"files"`
	Options map[string]string `json:"options"`
}

func CreateJob(c *fiber.Ctx) error {
	var body CreateJobRequest

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	if len(body.FileURLs) == 0 {
		return c.Status(400).JSON(fiber.Map{"error": "No files provided"})
	}

	jobID := uuid.New().String()

	redis.SaveJob(jobID, "queued")

	queueURL := jobs.QueueForTool(body.Tool)
	if queueURL == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid tool"})
	}


	job := map[string]interface{}{
		"id":    jobID,
		"tool":  body.Tool,
		"files": body.FileURLs,
		"status": "queued",
		"options": body.Options,
	}

	queue.PushJob(queueURL, job)

	redis.SaveJob(jobID, "queued")



	return c.JSON(fiber.Map{
		"job_id": jobID,
		"status": "queued",
	})
}

func GetJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")

	status := redis.GetJobStatus(jobID)

	return c.JSON(fiber.Map{
		"job_id": jobID,
		"status": status,
	})
}


func GetJobResult(c *fiber.Ctx) error {
	jobID := c.Params("id")

	data := redis.GetResult(jobID)
	if data == "" {
		return c.Status(404).JSON(fiber.Map{"error": "result not found"})
	}

	var urls []string

	// Try to parse JSON array
	if err := json.Unmarshal([]byte(data), &urls); err != nil {
		// If parsing fails → old single URL format
		return c.JSON(fiber.Map{
			"file_url": data,
		})
	}

	// If only one file
	if len(urls) == 1 {
		return c.JSON(fiber.Map{
			"file_url": urls[0],
		})
	}

	// Multiple files (e.g., split)
	return c.JSON(fiber.Map{
		"file_urls": urls,
	})
}




