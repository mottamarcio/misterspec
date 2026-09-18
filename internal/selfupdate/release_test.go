package selfupdate

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type fakeRoundTripper struct {
	status int
	body   string
}

func (f fakeRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: f.status,
		Body:       io.NopCloser(strings.NewReader(f.body)),
		Header:     make(http.Header),
	}, nil
}

const canned = `{
  "tag_name": "v1.2.0",
  "assets": [
    {"name": "misterspec-linux-amd64", "browser_download_url": "https://example.com/misterspec-linux-amd64"},
    {"name": "misterspec-darwin-arm64", "browser_download_url": "https://example.com/misterspec-darwin-arm64"},
    {"name": "misterspec-windows-amd64.exe", "browser_download_url": "https://example.com/misterspec-windows-amd64.exe"},
    {"name": "SHA256SUMS", "browser_download_url": "https://example.com/SHA256SUMS"}
  ]
}`

func TestLatestRelease(t *testing.T) {
	client := &http.Client{Transport: fakeRoundTripper{status: 200, body: canned}}
	rel, err := LatestRelease(context.Background(), client)
	if err != nil {
		t.Fatalf("LatestRelease() error = %v", err)
	}
	if rel.TagName != "v1.2.0" {
		t.Errorf("TagName = %q, want %q", rel.TagName, "v1.2.0")
	}
	if len(rel.Assets) != 4 {
		t.Errorf("len(Assets) = %d, want 4", len(rel.Assets))
	}
}

func TestLatestRelease_Failure(t *testing.T) {
	client := &http.Client{Transport: fakeRoundTripper{status: 500, body: "boom"}}
	if _, err := LatestRelease(context.Background(), client); err == nil {
		t.Fatal("LatestRelease() error = nil, want non-nil on a failed request")
	}
}

func TestSelectAsset(t *testing.T) {
	client := &http.Client{Transport: fakeRoundTripper{status: 200, body: canned}}
	rel, _ := LatestRelease(context.Background(), client)

	asset, ok := SelectAsset(rel, "linux", "amd64")
	if !ok || asset.Name != "misterspec-linux-amd64" {
		t.Errorf("SelectAsset(linux, amd64) = %+v, %v", asset, ok)
	}

	asset, ok = SelectAsset(rel, "windows", "amd64")
	if !ok || asset.Name != "misterspec-windows-amd64.exe" {
		t.Errorf("SelectAsset(windows, amd64) = %+v, %v", asset, ok)
	}

	if _, ok := SelectAsset(rel, "plan9", "amd64"); ok {
		t.Errorf("SelectAsset(plan9, amd64) ok = true, want false (no compatible asset)")
	}
}
