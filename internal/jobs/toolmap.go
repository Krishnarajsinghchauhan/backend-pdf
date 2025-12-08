package jobs

import "os"

func QueueForTool(tool string) string {
	switch tool {

	// Worker A — PDF operations
	case "merge", "split", "compress", "rotate", "delete-pages", "reorder", "protect", "unlock":
		return os.Getenv("PDF_QUEUE_URL")

	// Worker B — Image <-> PDF conversions
	case "jpg-to-pdf", "png-to-pdf", "pdf-to-jpg", "pdf-to-png":
		return os.Getenv("IMAGE_QUEUE_URL")

	// Worker C — Office conversions
	case "word-to-pdf", "excel-to-pdf", "ppt-to-pdf",
		"pdf-to-word", "pdf-to-excel", "pdf-to-ppt":
		return os.Getenv("OFFICE_QUEUE_URL")

	// Worker D — OCR tools
	case "ocr", "image-to-text", "scanned-enhance":
		return os.Getenv("OCR_QUEUE_URL")

	// Worker E — PDF Editor tools
	case "watermark", "page-numbers", "header-footer", "edit":
		return os.Getenv("EDITOR_QUEUE_URL")

	// Worker F — eSign tools
	case "esign":
		return os.Getenv("ESIGN_QUEUE_URL")

	// Worker G — Combine docs IMG + PDF + Office
	case "combine":
		return os.Getenv("COMBINE_QUEUE_URL")

	default:
		return ""
	}
}
