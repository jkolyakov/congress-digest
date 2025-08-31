package congress

import "time"

// Domain model for a Daily Digest entry. Adjust fields as per API schema.
type DailyDigest struct {
	Congress    string
	Issue       string
	PublishDate time.Time
	PDFUrl      string
	// Add fields you actually need.
}

// Structs matching the actual API response
type congressionalRecordResponse struct {
	Results struct {
		Issues []struct {
			Congress    string `json:"Congress"`
			Issue       string `json:"Issue"`
			PublishDate string `json:"PublishDate"`
			Links       struct {
				Digest struct {
					PDF []struct {
						Url string `json:"Url"`
					} `json:"PDF"`
				} `json:"Digest"`
			} `json:"Links"`
		} `json:"Issues"`
	} `json:"Results"`
}
