package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"net/http"

	"github.com/jkolyakov/congress-digest/internal/config"
	"github.com/jkolyakov/congress-digest/internal/congress"
	"github.com/jkolyakov/congress-digest/internal/printing"
)

func main() {
	_ = godotenv.Load() // optional; don't fail hard

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	httpClient := &http.Client{Timeout: cfg.HTTPTimeout}
	api := congress.NewClient(cfg.BaseURL, cfg.APIKey, httpClient)

	// Context with timeout to bound network operation
	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTPTimeout)
	defer cancel()

	d, err := api.LatestDailyDigest(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
		os.Exit(1)
	}

	// Print the result to stdout
	fmt.Println(printing.FormatDigest(d))

	// Print PDF as text
	daily_digest, err := api.DailyDigestText(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "fetch error: %v\n", err)
		os.Exit(1)
	}

	// Print the result to stdout
	fmt.Println(printing.FormatDigestText(daily_digest))

}
