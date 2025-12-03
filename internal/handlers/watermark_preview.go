package handlers

import (
    "backend/internal/s3"
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

    // temp files
    inputPDF := fmt.Sprintf("/tmp/preview_%d.pdf", time.Now().UnixNano())
    outputPNG := fmt.Sprintf("/tmp/preview_out_%d.png", time.Now().UnixNano())

    // 1. download PDF
    if err := download(req.PDFUrl, inputPDF); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Download PDF failed")
    }

    // 2. extract options
    text := safeString(req.Options["text"], "WATERMARK")
    color := safeString(req.Options["color"], "#000000")
    opacity := safeString(req.Options["opacity"], "0.25")
    angle := safeString(req.Options["angle"], "0")
    fontSize := safeString(req.Options["fontSize"], "60")

    // 3. generate watermark preview (page 1)
    cmdStr := fmt.Sprintf(
        `convert "%s[0]" -fill "%s" -gravity center -pointsize %s -annotate %s "%s" -alpha set -channel A -evaluate multiply %s "%s"`,
        inputPDF, color, fontSize, angle, text, opacity, outputPNG,
    )

    cmd := exec.Command("bash", "-c", cmdStr)
    if err := cmd.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Preview generation failed")
    }

    // 4. Upload preview to S3
    file, err := os.Open(outputPNG)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Open PNG failed")
    }
    defer file.Close()

    key := fmt.Sprintf("previews/%d.png", time.Now().UnixNano())

    // Upload using your existing S3 helper
    url, err := s3.UploadFile(file, key)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Upload preview failed")
    }

    return c.JSON(fiber.Map{
        "preview_url": url,
    })
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

    out, err := os.Create(output)
    if err != nil {
        return err
    }
    defer out.Close()

    _, err = io.Copy(out, resp.Body)
    return err
}

func safeString(v interface{}, def string) string {
    if s, ok := v.(string); ok {
        return s
    }
    return def
}
