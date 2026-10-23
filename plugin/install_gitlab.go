package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
)

func (c *InstallConfig) fetchGitLabReleaseAssets(ctx context.Context, client *http.Client) (map[string]string, error) {
	project := url.PathEscape(c.SourceOwner + "/" + c.SourceRepo)
	endpoint := fmt.Sprintf("https://%s/api/v4/projects/%s/releases/%s", c.SourceHost, project, url.PathEscape(c.TagName()))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitLab API returned %s", resp.Status)
	}
	var release struct {
		Assets struct {
			Links []struct {
				Name string `json:"name"`
				URL  string `json:"url"`
			} `json:"links"`
		} `json:"assets"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, err
	}
	assets := make(map[string]string, len(release.Assets.Links))
	for _, asset := range release.Assets.Links {
		if _, exists := assets[asset.Name]; exists {
			return nil, fmt.Errorf("duplicate asset name in GitLab release: %s", asset.Name)
		}
		assets[asset.Name] = asset.URL
	}
	return assets, nil
}

func downloadGitLabAsset(ctx context.Context, client *http.Client, assetURL string) (*os.File, error) {
	if assetURL == "" {
		return nil, fmt.Errorf("file not found in the GitLab release. Does the release contain an asset link with the correct name ?")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, assetURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("release asset download returned %s", resp.Status)
	}
	return downloadToTempFile(resp.Body)
}

func newGitLabClient(config *InstallConfig) *http.Client {
	transport := &gitLabTransport{
		RoundTripper: http.DefaultTransport,
		host:         config.SourceHost,
		header:       "PRIVATE-TOKEN",
		token:        config.getToken("GITLAB_TOKEN"),
	}
	if transport.token == "" {
		// GitLab skips empty host tokens; GitHub retains its existing empty override.
		transport.token = os.Getenv("GITLAB_TOKEN")
	}
	if transport.token == "" {
		ciURL, err := url.Parse(os.Getenv("CI_SERVER_URL"))
		if err == nil && ciURL.Scheme == "https" && strings.EqualFold(ciURL.Host, config.SourceHost) {
			transport.header = "JOB-TOKEN"
			transport.token = os.Getenv("CI_JOB_TOKEN")
		}
	}
	return &http.Client{Transport: transport}
}

type gitLabTransport struct {
	http.RoundTripper
	host   string
	header string
	token  string
}

func (t *gitLabTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Scheme != "https" || req.URL.User != nil {
		return nil, fmt.Errorf("GitLab release requests require HTTPS URLs without user information")
	}
	if t.token != "" && strings.EqualFold(req.URL.Host, t.host) {
		// Authenticate only this origin, without copying credentials into redirects.
		req = req.Clone(req.Context())
		req.Header.Set(t.header, t.token)
	}
	return t.RoundTripper.RoundTrip(req)
}
