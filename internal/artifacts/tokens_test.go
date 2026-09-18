package artifacts_test

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/artifacts"
)

func TestEstimateTokens_Deterministic(t *testing.T) {
	text := "The quick brown fox jumps over the lazy dog."

	first := artifacts.EstimateTokens(text)
	second := artifacts.EstimateTokens(text)

	if first != second {
		t.Errorf("EstimateTokens() = %d then %d, want identical", first, second)
	}
}

func TestEstimateTokens_LongerTextYieldsLargerEstimate(t *testing.T) {
	short := "Short."
	long := "This is a considerably longer piece of text than the short one above."

	if artifacts.EstimateTokens(long) <= artifacts.EstimateTokens(short) {
		t.Errorf("EstimateTokens(long) = %d, want it larger than EstimateTokens(short) = %d",
			artifacts.EstimateTokens(long), artifacts.EstimateTokens(short))
	}
}

func TestEstimateTokens_EmptyStringIsZero(t *testing.T) {
	if got := artifacts.EstimateTokens(""); got != 0 {
		t.Errorf("EstimateTokens(\"\") = %d, want 0", got)
	}
}

func TestDefaultEstimator_MatchesEstimateTokens(t *testing.T) {
	text := "Some representative chunk content."
	var e artifacts.Estimator = artifacts.DefaultEstimator{}

	if got, want := e.Estimate(text), artifacts.EstimateTokens(text); got != want {
		t.Errorf("DefaultEstimator{}.Estimate() = %d, want %d (EstimateTokens)", got, want)
	}
}

func TestDefaultEstimator_NameIsNonEmptyAndStable(t *testing.T) {
	var e artifacts.Estimator = artifacts.DefaultEstimator{}

	if e.Name() == "" {
		t.Fatal("DefaultEstimator{}.Name() is empty, want a non-empty, stable identifier (FR-001)")
	}
	if got, want := e.Name(), "default"; got != want {
		t.Errorf("DefaultEstimator{}.Name() = %q, want %q", got, want)
	}
}
