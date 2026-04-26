package v1

import (
	"fmt"
	"os/exec"
	"strings"
)

// extractText uses pdftotext (poppler-utils) to produce layout-preserving text
// from a PDF. The -layout flag keeps column positions intact, which is essential
// for parsing the fixed-width Piraeus Bank statement tables.
func extractText(pdfPath string) (string, error) {
	out, err := exec.Command("pdftotext", "-layout", "-enc", "UTF-8", pdfPath, "-").Output()
	if err != nil {
		return "", fmt.Errorf("pdftotext failed (ensure poppler-utils is installed): %w", err)
	}
	return string(out), nil
}

func splitPages(text string) []string {
	return strings.Split(text, "\f")
}
