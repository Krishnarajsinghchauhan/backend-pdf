package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"

	"backend/internal/config"
	"backend/internal/handlers"
	"backend/internal/queue"
	"backend/internal/redis"
)

func main() {
	// Load environment
	config.LoadEnv()

	redis.Init()
	queue.Init()

	app := fiber.New()

	// CORS for frontend → backend communication
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Root
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "OK"})
	})

	// Upload route
	app.Post("/upload/create-url", handlers.CreateUploadURL)

	// Job routes
	app.Post("/job/create", handlers.CreateJob)
	app.Get("/job/status/:id", handlers.GetJobStatus)
	app.Get("/job/result/:id", handlers.GetJobResult)

	log.Println("API Gateway running on port", os.Getenv("PORT"))
	app.Listen(":" + os.Getenv("PORT"))
}
