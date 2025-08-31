// printing/digest.go

package printing

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/jkolyakov/congress-digest/internal/congress"
)

// FormatDigest prints a summary of the Congressional Record issue (metadata only).
func FormatDigest(d congress.DailyDigest) string {
	date := d.PublishDate.Format("2006-01-02")
	if d.PublishDate.IsZero() {
		date = "(unknown date)"
	}

	out := "# Congress.gov Daily Digest\n\n"
	out += fmt.Sprintf("**Issue:** %s\n", d.Issue)
	out += fmt.Sprintf("**PublishDate:** %s\n", date)
	if d.PDFUrl != "" {
		out += fmt.Sprintf("**PDFUrl:** %s\n", d.PDFUrl)
	}
	return out
}

// FormatDigestText cleans and formats the raw text extracted from pdftotext for easier reading.
func FormatDigestText(raw string) string {
	lines := strings.Split(raw, "\n")
	var cleaned []string

	// Regexes for common artifacts and syntax warnings to skip
	artifactPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^VerDate`),
		regexp.MustCompile(`^DMWilson on DSK`),
		regexp.MustCompile(`^E:\S+`),
		regexp.MustCompile(`^\s*Page \w+`),
		regexp.MustCompile(`^\s*D\d+\s*$`),     // e.g. D847
		regexp.MustCompile(`^\s*$`),            // blank lines
		regexp.MustCompile(`^Syntax Warning:`), // pdftotext syntax warnings
		regexp.MustCompile(`^[A-Z ]{2,}$`),     // lines with only capital letters (often artifacts)
		regexp.MustCompile(`^E PL$`),           // artifact from PDF
		regexp.MustCompile(`^M$`),              // artifact from PDF
		regexp.MustCompile(`^UR$`),             // artifact from PDF
		regexp.MustCompile(`^IB\s*NU$`),        // artifact from PDF
		regexp.MustCompile(`^U$`),              // artifact from PDF
		regexp.MustCompile(`^S$`),              // artifact from PDF
	}

	stopAt := "Congressional Record"
	stopFound := false

	for _, line := range lines {
		if stopFound {
			break
		}
		trimmed := strings.TrimSpace(line)
		if trimmed == stopAt {
			stopFound = true
			break
		}
		skip := false
		for _, re := range artifactPatterns {
			if re.MatchString(line) {
				skip = true
				break
			}
		}
		if skip {
			continue
		}
		// Collapse multiple spaces to one, trim
		cleaned = append(cleaned, trimmed)
	}

	// Remove consecutive blank lines
	var result []string
	lastBlank := false
	for _, line := range cleaned {
		if line == "" {
			if !lastBlank {
				result = append(result, "")
			}
			lastBlank = true
		} else {
			result = append(result, line)
			lastBlank = false
		}
	}

	// Optionally, add markdown headers for major sections
	sectionHeaders := []string{
		"Daily Digest",
		"Senate",
		"House of Representatives",
		"Joint Meetings",
		"Committee Meetings",
		"Extensions of Remarks",
		"Next Meeting of the SENATE",
		"Next Meeting of the HOUSE OF REPRESENTATIVES",
	}
	for i, line := range result {
		for _, header := range sectionHeaders {
			if strings.HasPrefix(line, header) {
				result[i] = "## " + line
			}
		}
	}

	return strings.Join(result, "\n")
}

// Small helper for deterministic timeouts from main.
func WithDefaultTimeout(now time.Time, d time.Duration) time.Time { return now.Add(d) }
