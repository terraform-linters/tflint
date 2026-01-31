package versioncheck

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-github/v81/github"
)

func TestFetchLatestReleaseWithClient(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		response    interface{}
		wantErr     bool
		wantVersion string
		checkReq    func(*testing.T, *http.Request)
	}{
		{
			name:       "successful fetch",
			statusCode: http.StatusOK,
			response: &github.RepositoryRelease{
				TagName: github.Ptr("v0.60.0"),
			},
			wantVersion: "v0.60.0",
		},
		{
			name:       "missing tag name",
			statusCode: http.StatusOK,
			response: &github.RepositoryRelease{
				TagName: nil,
			},
			wantErr: true,
		},
		{
			name:       "rate limit error",
			statusCode: http.StatusForbidden,
			response: map[string]interface{}{
				"message": "API rate limit exceeded",
			},
			wantErr: true,
		},
		{
			name:       "not found error",
			statusCode: http.StatusNotFound,
			response: map[string]interface{}{
				"message": "Not Found",
			},
			wantErr: true,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			response: map[string]interface{}{
				"message": "Internal Server Error",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkReq != nil {
					tt.checkReq(t, r)
				}

				w.WriteHeader(tt.statusCode)
				if tt.response != nil {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			// Create a custom client that points to our test server
			testClient := &http.Client{
				Transport: &testTransport{
					baseURL: server.URL,
				},
			}

			ctx := context.Background()
			got, err := fetchLatestReleaseWithClient(ctx, testClient)

			if (err != nil) != tt.wantErr {
				t.Errorf("fetchLatestReleaseWithClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && got != tt.wantVersion {
				t.Errorf("fetchLatestReleaseWithClient() = %v, want %v", got, tt.wantVersion)
			}
		})
	}
}

// testTransport is a custom RoundTripper that redirects requests to a test server
type testTransport struct {
	baseURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Rewrite the request URL to point to our test server
	testURL, err := url.Parse(t.baseURL)
	if err != nil {
		return nil, err
	}

	req.URL.Scheme = testURL.Scheme
	req.URL.Host = testURL.Host

	return http.DefaultTransport.RoundTrip(req)
}
