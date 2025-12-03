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

    inputPDF := fmt.Sprintf("/tmp/preview_%d.pdf", time.Now().UnixNano())
    outputPNG := fmt.Sprintf("/tmp/preview_out_%d.png", time.Now().UnixNano())

    // download PDF
    if err := download(req.PDFUrl, inputPDF); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Download PDF failed")
    }

    text := safeString(req.Options["text"], "WATERMARK")
    color := safeString(req.Options["color"], "#000000")
    opacity := safeString(req.Options["opacity"], "0.25")
    angle := safeString(req.Options["angle"], "0")
    fontSize := safeString(req.Options["fontSize"], "60")

    // REAL command (ImageMagick 6)
    cmdStr := fmt.Sprintf(
        `convert "%s[0]" -fill "%s" -gravity center -pointsize %s -annotate %s "%s" -alpha set -channel A -evaluate multiply %s "%s"`,
        inputPDF, color, fontSize, angle, text, opacity, outputPNG,
    )

    cmd := exec.Command("bash", "-c", cmdStr)
    if err := cmd.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Preview generation failed")
    }

    // TODO: upload to S3
    url := "https://pixelpdf.in/previews/" + fmt.Sprintf("%d.png", time.Now().UnixNano())

    return c.JSON(fiber.Map{"preview_url": url})
}
