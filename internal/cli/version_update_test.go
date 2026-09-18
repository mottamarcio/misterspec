package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/selfupdate"
)

// runRootCmd executes newRootCmd() directly, mirroring runInitCmd's own
// convention (init_test.go).
func runRootCmd(args []string) (output string, exitCode int) {
	cmd := newRootCmd()
	cmd.SetArgs(args)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	if err == nil {
		return buf.String(), 0
	}
	var exitErr *internalcmd.ExitCodeError
	if errors.As(err, &exitErr) {
		return buf.String(), exitErr.Code
	}
	return buf.String(), -1
}

func TestVersionFlag_ReportsEmbeddedVersion(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.2.0"

	output, exitCode := runRootCmd([]string{"--version"})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK      bool   `json:"ok"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decoding output %q: %v", output, err)
	}
	if !decoded.OK || decoded.Version != "v1.2.0" {
		t.Errorf("decoded = %+v, want ok=true version=v1.2.0", decoded)
	}
}

func TestVersionFlag_DevelopmentBuild(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = ""

	output, exitCode := runRootCmd([]string{"--version"})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}
	if !strings.Contains(output, `"development build"`) {
		t.Errorf("output = %q, want it to contain %q", output, "development build")
	}
}

// fakeRT is a scriptable http.RoundTripper for --update's own tests —
// keyed by exact URL so a single fake can serve both the release-lookup
// call and any asset-download calls in the same test.
type fakeRT struct {
	responses map[string]fakeResponse
}

type fakeResponse struct {
	status int
	body   string
	err    error
}

func (f fakeRT) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, ok := f.responses[req.URL.String()]
	if !ok {
		return &http.Response{StatusCode: 404, Body: io.NopCloser(strings.NewReader("not found")), Header: make(http.Header)}, nil
	}
	if resp.err != nil {
		return nil, resp.err
	}
	return &http.Response{StatusCode: resp.status, Body: io.NopCloser(strings.NewReader(resp.body)), Header: make(http.Header)}, nil
}

func TestUpdateFlag_UpToDate(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.2.0"

	httpClient = &http.Client{Transport: fakeRT{responses: map[string]fakeResponse{
		selfupdate.ReleasesAPIURL: {status: 200, body: `{"tag_name":"v1.2.0","assets":[]}`},
	}}}
	t.Cleanup(func() { httpClient = http.DefaultClient })

	output, exitCode := runRootCmd([]string{"--update"})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		OK     bool `json:"ok"`
		Update struct {
			Status string `json:"status"`
		} `json:"update"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decoding output %q: %v", output, err)
	}
	if decoded.Update.Status != "up_to_date" {
		t.Errorf("status = %q, want %q", decoded.Update.Status, "up_to_date")
	}
}

func TestUpdateFlag_DeclinedConfirmation(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.1.5-alpha"

	httpClient = &http.Client{Transport: fakeRT{responses: map[string]fakeResponse{
		selfupdate.ReleasesAPIURL: {status: 200, body: `{"tag_name":"v1.2.0","assets":[]}`},
	}}}
	t.Cleanup(func() { httpClient = http.DefaultClient })

	confirmUpdate = func(current, latest string) bool { return false }
	t.Cleanup(func() { confirmUpdate = defaultConfirmUpdate })

	output, exitCode := runRootCmd([]string{"--update"})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Update struct {
			Status string `json:"status"`
		} `json:"update"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decoding output %q: %v", output, err)
	}
	if decoded.Update.Status != "declined" {
		t.Errorf("status = %q, want %q", decoded.Update.Status, "declined")
	}
}

func TestUpdateFlag_NoCompatibleAsset(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.1.5-alpha"

	httpClient = &http.Client{Transport: fakeRT{responses: map[string]fakeResponse{
		selfupdate.ReleasesAPIURL: {status: 200, body: `{"tag_name":"v1.2.0","assets":[]}`},
	}}}
	t.Cleanup(func() { httpClient = http.DefaultClient })
	confirmUpdate = func(current, latest string) bool { return true }
	t.Cleanup(func() { confirmUpdate = defaultConfirmUpdate })

	_, exitCode := runRootCmd([]string{"--update"})
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want a non-zero exit for no_compatible_asset")
	}
}

func TestUpdateFlag_ChecksumMismatchLeavesEverythingUnchanged(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.1.5-alpha"

	dir := t.TempDir()
	exePath := filepath.Join(dir, "misterspec")
	if err := os.WriteFile(exePath, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}

	assetName := "misterspec-" + goosOverride + "-" + goarchOverride
	if goosOverride == "windows" {
		assetName += ".exe"
	}
	assetURL := "https://example.com/" + assetName
	sumsURL := "https://example.com/SHA256SUMS"

	httpClient = &http.Client{Transport: fakeRT{responses: map[string]fakeResponse{
		selfupdate.ReleasesAPIURL: {status: 200, body: `{"tag_name":"v1.2.0","assets":[` +
			`{"name":"` + assetName + `","browser_download_url":"` + assetURL + `"},` +
			`{"name":"SHA256SUMS","browser_download_url":"` + sumsURL + `"}` +
			`]}`},
		assetURL: {status: 200, body: "new-binary-content"},
		sumsURL:  {status: 200, body: "0000000000000000000000000000000000000000000000000000000000000000  " + assetName + "\n"},
	}}}
	t.Cleanup(func() { httpClient = http.DefaultClient })
	confirmUpdate = func(current, latest string) bool { return true }
	t.Cleanup(func() { confirmUpdate = defaultConfirmUpdate })

	executablePath = func() (string, error) { return exePath, nil }
	t.Cleanup(func() { executablePath = os.Executable })

	_, exitCode := runRootCmd([]string{"--update"})
	if exitCode == 0 {
		t.Fatalf("exitCode = 0, want a non-zero exit for checksum_mismatch")
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "original" {
		t.Errorf("exePath content = %q, want it untouched (%q)", data, "original")
	}
}

func TestUpdateFlag_SuccessfulUpdateNoProject(t *testing.T) {
	t.Cleanup(func() { selfupdate.Version = "" })
	selfupdate.Version = "v1.1.5-alpha"

	dir := t.TempDir()
	exePath := filepath.Join(dir, "misterspec")
	if err := os.WriteFile(exePath, []byte("original"), 0o755); err != nil {
		t.Fatal(err)
	}

	assetName := "misterspec-" + goosOverride + "-" + goarchOverride
	if goosOverride == "windows" {
		assetName += ".exe"
	}
	assetURL := "https://example.com/" + assetName
	sumsURL := "https://example.com/SHA256SUMS"
	content := "new-binary-content"
	sum := sha256Hex(content)

	httpClient = &http.Client{Transport: fakeRT{responses: map[string]fakeResponse{
		selfupdate.ReleasesAPIURL: {status: 200, body: `{"tag_name":"v1.2.0","assets":[` +
			`{"name":"` + assetName + `","browser_download_url":"` + assetURL + `"},` +
			`{"name":"SHA256SUMS","browser_download_url":"` + sumsURL + `"}` +
			`]}`},
		assetURL: {status: 200, body: content},
		sumsURL:  {status: 200, body: sum + "  " + assetName + "\n"},
	}}}
	t.Cleanup(func() { httpClient = http.DefaultClient })
	confirmUpdate = func(current, latest string) bool { return true }
	t.Cleanup(func() { confirmUpdate = defaultConfirmUpdate })

	executablePath = func() (string, error) { return exePath, nil }
	t.Cleanup(func() { executablePath = os.Executable })

	updateWorkDir = func() (string, error) { return t.TempDir(), nil } // no .misterspec/install.json here
	t.Cleanup(func() { updateWorkDir = os.Getwd })

	output, exitCode := runRootCmd([]string{"--update"})
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0 (output: %s)", exitCode, output)
	}

	var decoded struct {
		Update struct {
			Status          string `json:"status"`
			SkillsRefreshed bool   `json:"skills_refreshed"`
		} `json:"update"`
	}
	if err := json.Unmarshal([]byte(output), &decoded); err != nil {
		t.Fatalf("decoding output %q: %v", output, err)
	}
	if decoded.Update.Status != "updated" {
		t.Errorf("status = %q, want %q", decoded.Update.Status, "updated")
	}
	if decoded.Update.SkillsRefreshed {
		t.Errorf("skills_refreshed = true, want false (no project in this test's own working dir)")
	}

	data, err := os.ReadFile(exePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != content {
		t.Errorf("exePath content = %q, want %q", data, content)
	}
}

func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}
