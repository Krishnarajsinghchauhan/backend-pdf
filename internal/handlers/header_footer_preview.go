package handlers

import (
    "backend/internal/s3uploader"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "time"

    "github.com/gofiber/fiber/v2"
)

type HeaderFooterPreviewReq struct {
    PDFUrl string                 `json:"pdfUrl"`
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

    inputPDF := fmt.Sprintf("/tmp/hf_preview_%d.pdf", time.Now().UnixNano())
    outputPNG := fmt.Sprintf("/tmp/hf_preview_out_%d.png", time.Now().UnixNano())
    layerPNG := fmt.Sprintf("/tmp/hf_layer_%d.png", time.Now().UnixNano())

    // Download PDF
    if err := download(req.PDFUrl, inputPDF); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Download PDF failed")
    }

    // Extract options
    header := safeString(req.Options["header"], "")
    footer := safeString(req.Options["footer"], "")
    fontSize := safeString(req.Options["fontSize"], "40")
    color := safeString(req.Options["color"], "#000000")
    align := safeString(req.Options["align"], "center")
    marginTop := safeString(req.Options["marginTop"], "80")
    marginBottom := safeString(req.Options["marginBottom"], "80")

    // Create header/footer overlay (PNG)
    cmdLayer := exec.Command("bash", "-c",
        fmt.Sprintf(`
convert -size 2480x3508 xc:none \
  -gravity north -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  -gravity south -pointsize %s -fill "%s" -annotate +0+%s "%s" \
  "%s"
`, fontSize, color, marginTop, header, fontSize, color, marginBottom, footer, layerPNG))

    if err := cmdLayer.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Layer creation failed")
    }

    // Build preview on page 1
    cmdPrev := exec.Command("bash", "-c",
        fmt.Sprintf(`
convert "%s[0]" "%s" -gravity center -compose over -composite "%s"
`, inputPDF, layerPNG, outputPNG))

    if err := cmdPrev.Run(); err != nil {
        return fiber.NewError(fiber.StatusInternalServerError, "Preview generation failed")
    }

    // Upload preview PNG
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

func safeString(v interface{}, def string) string {
    if s, ok := v.(string); ok {
        return s
    }
    return def
}

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
