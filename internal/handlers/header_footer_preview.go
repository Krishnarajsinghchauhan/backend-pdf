package handlers

import (
    "backend/internal/s3uploader"
    "fmt"
    "os"
    "os/exec"
    "time"

    "github.com/gofiber/fiber/v2"
)

type HeaderFooterPreviewReq struct {
    PDFUrl  string                 `json:"pdfUrl"`
    Options map[string]interface{} `json:"options"`
}

func HeaderFooterPreview(c *fiber.Ctx) error {
    var req HeaderFooterPreviewReq
    if err := c.BodyParser(&req); err != nil {
        return fiber.NewError(fiber.StatusBadRequest, "Invalid payload")
    }

    if req.PDFUrl == "" {
        return fiber.NewError(fiber.StatusBadRequest, "Missing pdfUrl")
    }

    // Temp files
    inputPDF := fmt.Sprintf("/tmp/hf_preview_%d.pdf", time.Now().UnixNano())
    outputPNG := fmt.Sprintf("/tmp/hf_preview_out_%d.png", time.Now().UnixNano())
    layerPNG := fmt.Sprintf("/tmp/hf_layer_%d.png", time.Now().UnixNano())

    // 🔹 Download source PDF
    if err := download(req.PDFUrl, inputPDF); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Download PDF failed")
    }

    // Extract fields
    header := safeString(req.Options["header"], "")
    footer := safeString(req.Options["footer"], "")
    fontSize := safeString(req.Options["fontSize"], "40")
    color := safeString(req.Options["color"], "#000000")
    align := safeString(req.Options["align"], "center")
    marginTop := safeString(req.Options["marginTop"], "80")
    marginBottom := safeString(req.Options["marginBottom"], "80")

    // Map alignment to ImageMagick gravity
    gravity := map[string]string{
        "left":      "west",
        "right":     "east",
        "center":    "center",
        "top":       "north",
        "bottom":    "south",
        "topleft":   "northwest",
        "topright":  "northeast",
        "bottomleft": "southwest",
        "bottomright": "southeast",
    }[align]

    if gravity == "" {
        gravity = "center"
    }

    // 🔹 Build header/footer overlay PNG
    cmdLayer := exec.Command("bash", "-c",
        fmt.Sprintf(`
convert -size 2480x3508 xc:none \
  -gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  -gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  "%s"
`, fontSize, color, marginTop, header,
           fontSize, color, marginBottom, footer,
           layerPNG))

    if err := cmdLayer.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Layer creation failed")
    }

    // 🔹 Build preview using only page 1
    cmdPrev := exec.Command("bash", "-c",
        fmt.Sprintf(`
convert "%s[0]" "%s" -gravity %s -compose over -composite "%s"
`, inputPDF, layerPNG, gravity, outputPNG))

    if err := cmdPrev.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Preview generation failed")
    }

    // 🔹 Upload preview to S3
    pngBytes, err := os.ReadFile(outputPNG)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Read PNG failed")
    }

    key := fmt.Sprintf("previews/hf_%d.png", time.Now().UnixNano())
    url, err := s3uploader.UploadPublicFile(pngBytes, key)
    if err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Upload preview failed")
    }

    return c.JSON(fiber.Map{
        "preview_url": url,
    })
}
