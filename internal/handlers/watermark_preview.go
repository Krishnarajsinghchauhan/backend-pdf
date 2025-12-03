package handlers

import (
	"fmt"

	"os/exec"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Request structure
type PreviewRequest struct {
	PDFUrl  string                 `json:"pdfUrl"`
	Options map[string]interface{} `json:"options"`
}

func WatermarkPreview(c *fiber.Ctx) error {
	var req PreviewRequest

	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if req.PDFUrl == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing pdfUrl")
	}

	// Temp file paths
	inputPath := fmt.Sprintf("/tmp/input_%d.pdf", time.Now().UnixNano())
	outputPath := fmt.Sprintf("/tmp/preview_%d.png", time.Now().UnixNano())

	// 1) Download PDF to server
	err := downloadFile(req.PDFUrl, inputPath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "PDF download failed")
	}

	// 2) Apply watermark on FIRST PAGE ONLY
	text := req.Options["text"].(string)
	color := req.Options["color"].(string)
	opacity := req.Options["opacity"].(string)
	angle := req.Options["angle"].(string)
	fontSize := req.Options["fontSize"].(string)

	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(
			`magick convert -density 150 "%s[0]" \
        -fill "%s" -gravity center -pointsize %s \
        -annotate %s "%s" \
        -alpha set -channel A -evaluate multiply %s \
        "%s"`,
			inputPath, color, fontSize, angle, text, opacity, outputPath,
		),
	)

	err = cmd.Run()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Watermark preview failed")
	}

	// 3) Upload preview to S3
	previewURL, err := uploadToS3(outputPath)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Preview upload failed")
	}

	return c.JSON(fiber.Map{
		"preview_url": previewURL,
	})
}

// ------------------------------
// Download Helper
// ------------------------------
func downloadFile(url, dest string) error {
	cmd := exec.Command("bash", "-c", fmt.Sprintf(`curl -s -o "%s" "%s"`, dest, url))
	return cmd.Run()
}

// ------------------------------
// Upload Helper (same as your worker)
// ------------------------------
func uploadToS3(path string) (string, error) {
	// Implement the same upload code you already use in workers
	return "", nil // ← You already have working upload code; reuse it.
}
