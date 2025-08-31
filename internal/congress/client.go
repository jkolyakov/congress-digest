// congress.go

package congress

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"
)

// Reader interface (Dependency Inversion) for the API.
type DigestReader interface {
	LatestDailyDigest(ctx context.Context) (DailyDigest, error)
	DailyDigestText(ctx context.Context) (string, error)
}

type Client struct {
	baseURL string
	apiKey  string
	http    *http.Client
}

func NewClient(baseURL, apiKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		http:    httpClient,
	}
}

func (c *Client) LatestDailyDigest(ctx context.Context) (DailyDigest, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/congressional-record?api_key=%s", c.baseURL, c.apiKey),
		nil,
	)
	if err != nil {
		return DailyDigest{}, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return DailyDigest{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	// Read and print the response body for debugging
	// b, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
	// fmt.Printf("Response Body: %s\n", string(b))
	// resp.Body = io.NopCloser(bytes.NewBuffer(b)) // Reset body for decoding

	// if resp.StatusCode != http.StatusOK {
	// 	return DailyDigest{}, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(b))
	// }

	var wire congressionalRecordResponse
	if err := json.NewDecoder(resp.Body).Decode(&wire); err != nil {
		return DailyDigest{}, fmt.Errorf("decode response: %w", err)
	}
	if len(wire.Results.Issues) == 0 {
		return DailyDigest{}, errors.New("no issues found")
	}

	latest := wire.Results.Issues[0]
	var parsedDate time.Time
	if latest.PublishDate != "" {
		if t, err := time.Parse("2006-01-02", latest.PublishDate); err == nil {
			parsedDate = t
		}
	}

	var pdfUrl string
	if len(latest.Links.Digest.PDF) > 0 {
		pdfUrl = latest.Links.Digest.PDF[0].Url
	}

	return DailyDigest{
		Congress:    latest.Congress,
		Issue:       latest.Issue,
		PublishDate: parsedDate,
		PDFUrl:      pdfUrl,
	}, nil
}

// DailyDigestText downloads the latest daily digest PDF and extracts its text using pdftotext CLI.
func (c *Client) DailyDigestText(ctx context.Context) (string, error) {
	digest, err := c.LatestDailyDigest(ctx)
	if err != nil {
		return "", err
	}
	if digest.PDFUrl == "" {
		return "", errors.New("no PDF URL found for daily digest")
	}

	// Download the PDF to a temporary file
	resp, err := c.http.Get(digest.PDFUrl)
	if err != nil {
		return "", fmt.Errorf("download PDF: %w", err)
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", "daily-digest-*.pdf")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		return "", fmt.Errorf("save PDF: %w", err)
	}

	// Run pdftotext on the downloaded file
	args := []string{
		"-layout",
		"-nopgbrk",
		tmpFile.Name(),
		"-",
	}
	cmd := exec.CommandContext(ctx, "pdftotext", args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf // capture errors as well

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed: %v\nOutput: %s", err, buf.String())
	}

	return buf.String(), nil
}
