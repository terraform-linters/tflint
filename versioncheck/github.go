package versioncheck

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/google/go-github/v81/github"
	"golang.org/x/oauth2"
)

const (
	repoOwner = "terraform-linters"
	repoName  = "tflint"
)

// fetchLatestRelease fetches the latest release version from GitHub
func fetchLatestRelease(ctx context.Context) (string, error) {
	return fetchLatestReleaseWithClient(ctx, nil)
}

// fetchLatestReleaseWithClient fetches the latest release version from GitHub using a custom HTTP client
// If httpClient is nil, creates a default client with optional GITHUB_TOKEN authentication
func fetchLatestReleaseWithClient(ctx context.Context, httpClient *http.Client) (string, error) {
	// Create GitHub client with optional authentication
	if httpClient == nil {
		httpClient = &http.Client{Transport: http.DefaultTransport}
		if token := os.Getenv("GITHUB_TOKEN"); token != "" {
			httpClient = oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: token,
			}))
		}
	}
	client := github.NewClient(httpClient)

	log.Printf("[DEBUG] Fetching latest release from GitHub API")
	release, resp, err := client.Repositories.GetLatestRelease(ctx, repoOwner, repoName)
	if err != nil {
		// Check if it's a rate limit error
		if resp != nil && resp.StatusCode == http.StatusForbidden {
			if resp.Rate.Remaining == 0 {
				log.Printf("[ERROR] GitHub API rate limited, consider setting GITHUB_TOKEN")
			}
		}
		return "", fmt.Errorf("failed to fetch latest release: %w", err)
	}

	if release.TagName == nil {
		return "", fmt.Errorf("latest release has no tag name")
	}

	log.Printf("[DEBUG] Latest release: %s", *release.TagName)
	return *release.TagName, nil
}
