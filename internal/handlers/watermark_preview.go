package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gofiber/fiber/v2"
)

type PreviewRequest struct {
	PDFUrl  string                 `json:"pdfUrl"`
	Options map[string]interface{} `json:"options"`
}

func WatermarkPreview(c *fiber.Ctx) error {
	var req PreviewRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
	}

	if req.PDFUrl == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Missing pdfUrl")
	}

	// create temp files
	inputPDF := fmt.Sprintf("/tmp/preview_%d.pdf", time.Now().UnixNano())
	outputPNG := fmt.Sprintf("/tmp/preview_out_%d.png", time.Now().UnixNano())

	// 🟦 download pdf
	err := download(req.PDFUrl, inputPDF)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Download PDF failed")
	}

	// 🟦 extract options
	text := safeString(req.Options["text"], "WATERMARK")
	color := safeString(req.Options["color"], "#000000")
	opacity := safeString(req.Options["opacity"], "0.25")
	angle := safeString(req.Options["angle"], "0")
	fontSize := safeString(req.Options["fontSize"], "60")

	// 🟦 generate preview (only page 1)
	cmd := exec.Command("bash", "-c",
		fmt.Sprintf(
			`magick "%s[0]" -fill "%s" -gravity center -pointsize %s `+
				`-annotate %s "%s" -alpha set -channel A -evaluate multiply %s "%s"`,
			inputPDF, color, fontSize, angle, text, opacity, outputPNG,
		),
	)

	if err := cmd.Run(); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Preview generation failed")
	}

	// 🟦 upload to S3
	url, err := uploadPreviewToS3(outputPNG)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Upload preview failed")
	}

	return c.JSON(fiber.Map{"preview_url": url})
}

// ------------------------------
// Helpers
// ------------------------------

func download(url, output string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	f, err := os.Create(output)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func safeString(val interface{}, def string) string {
	if v, ok := val.(string); ok {
		return v
	}
	return def
}

// TODO: replace this with real S3 uploader
func uploadPreviewToS3(path string) (string, error) {
	// temp: serve directly for now
	id := time.Now().UnixNano()
	public := fmt.Sprintf("https://pixelpdf.in/previews/%d.png", id)

	// you must copy file to public CDN or S3 bucket
	// cp path → /var/www/previews/... OR use your AWS uploader

	return public, nil
}
