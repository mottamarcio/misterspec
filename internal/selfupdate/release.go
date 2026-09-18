package selfupdate

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// ReleasesAPIURL is the GitHub Releases API endpoint for this project's
// own repository (research.md), exported so cli tests can key a fake
// http.RoundTripper's canned responses by the exact URL LatestRelease
// requests.
const ReleasesAPIURL = "https://api.github.com/repos/mottamarcio/misterspec/releases/latest"

// Asset is one binary or checksum file published alongside a Release.
type Asset struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

// Release is the subset of GitHub's own release JSON this feature needs.
type Release struct {
	TagName string  `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

// LatestRelease fetches the latest published release from GitHub via the
// given *http.Client (its Transport is always injected by the caller —
// production code passes a real client, tests pass a fake
// http.RoundTripper, so no test ever makes a real network call;
// research.md). A non-2xx response or a body that fails to parse is
// reported as a plain error — callers map this to the "release check
// failed" outcome (FR-012).
func LatestRelease(ctx context.Context, client *http.Client) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, ReleasesAPIURL, nil)
	if err != nil {
		return Release{}, fmt.Errorf("%w: %v", ErrReleaseCheckFailed, err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := client.Do(req)
	if err != nil {
		return Release{}, fmt.Errorf("%w: %v", ErrReleaseCheckFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Release{}, fmt.Errorf("%w: HTTP %d", ErrReleaseCheckFailed, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Release{}, fmt.Errorf("%w: %v", ErrReleaseCheckFailed, err)
	}

	var rel Release
	if err := json.Unmarshal(body, &rel); err != nil {
		return Release{}, fmt.Errorf("%w: parsing release response: %v", ErrReleaseCheckFailed, err)
	}
	return rel, nil
}

// SelectAsset finds the Release's own asset matching goos/goarch, per
// release.yml's own existing naming convention
// (misterspec-<goos>-<goarch>[.exe]). ok is false when no asset matches
// the current platform — callers map that to "no compatible asset"
// (FR-012).
func SelectAsset(rel Release, goos, goarch string) (Asset, bool) {
	name := "misterspec-" + goos + "-" + goarch
	if goos == "windows" {
		name += ".exe"
	}
	for _, a := range rel.Assets {
		if a.Name == name {
			return a, true
		}
	}
	return Asset{}, false
}

// ChecksumsAsset finds the SHA256SUMS asset within a Release's own
// assets, if published (research.md).
func ChecksumsAsset(rel Release) (Asset, bool) {
	for _, a := range rel.Assets {
		if a.Name == "SHA256SUMS" {
			return a, true
		}
	}
	return Asset{}, false
}

// Download fetches an Asset's own content via the given *http.Client —
// the same injected-transport convention as LatestRelease.
func Download(ctx context.Context, client *http.Client, a Asset) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReleaseCheckFailed, err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrReleaseCheckFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%w: downloading %s failed: HTTP %d", ErrReleaseCheckFailed, a.Name, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
