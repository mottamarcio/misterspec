package contextengine

import (
	"strings"
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestPayloadTokens_RenderedUsesInjectedEstimator(t *testing.T) {
	rendered := "Some rendered Markdown context pack body."

	got := PayloadTokens(0, nil, rendered, stubEstimator{fixed: 42})
	if got != 42 {
		t.Errorf("PayloadTokens() = %d, want 42 (from the injected estimator, not artifacts.EstimateTokens)", got)
	}

	want := artifacts.EstimateTokens(rendered)
	got = PayloadTokens(0, nil, rendered, artifacts.DefaultEstimator{})
	if got != want {
		t.Errorf("PayloadTokens() with DefaultEstimator = %d, want %d (matching artifacts.EstimateTokens on the actually-rendered text — FR-004)", got, want)
	}
}

func TestPayloadTokens_PackageOverheadUsesInjectedEstimator(t *testing.T) {
	items := []PackageItem{
		{Path: "ai/specs/SPEC-014/spec.md", Heading: "Requirements", Fingerprint: "sha256:abc"},
	}

	got := PayloadTokens(10, items, "", stubEstimator{fixed: 3})
	if got != 10+3 {
		t.Errorf("PayloadTokens() = %d, want %d (contentTokens + one stub-estimated overhead call)", got, 10+3)
	}
}

func TestPayloadTokens_CorrespondsToActualDeliveredText(t *testing.T) {
	rendered := strings.Repeat("word ", 200)
	got := PayloadTokens(0, nil, rendered, artifacts.DefaultEstimator{})
	want := artifacts.DefaultEstimator{}.Estimate(rendered)
	if got != want {
		t.Errorf("PayloadTokens() = %d, want %d — must correspond to the text actually delivered, not a disconnected approximation (FR-004)", got, want)
	}
}
