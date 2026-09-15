package bootstrap_test

import (
	"path/filepath"
	"testing"

	"github.com/mottamarcio/misterspec/internal/bootstrap"
)

func TestVerify_SuccessfulBootstrapMatches(t *testing.T) {
	target := filepath.Join(t.TempDir(), "verified-project")
	if _, err := bootstrap.Bootstrap(target, "fake-agent", fixtureRegistry(), fixtureSkills()); err != nil {
		t.Fatalf("Bootstrap() unexpected error: %v", err)
	}

	result, err := bootstrap.Verify(target, "fake-agent")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if !result.Detected {
		t.Error("Verify() Detected = false, want true")
	}
	if result.InstalledAgent != "fake-agent" {
		t.Errorf("Verify() InstalledAgent = %q, want %q", result.InstalledAgent, "fake-agent")
	}
	if !result.AgentMatches {
		t.Error("Verify() AgentMatches = false, want true")
	}
}

func TestVerify_MismatchedAgentIsNotAnError(t *testing.T) {
	target := filepath.Join(t.TempDir(), "mismatched-project")
	if _, err := bootstrap.Bootstrap(target, "fake-agent", fixtureRegistry(), fixtureSkills()); err != nil {
		t.Fatalf("Bootstrap() unexpected error: %v", err)
	}

	result, err := bootstrap.Verify(target, "some-other-agent")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if !result.Detected {
		t.Error("Verify() Detected = false, want true — the project itself is still valid")
	}
	if result.AgentMatches {
		t.Error("Verify() AgentMatches = true, want false for a mismatched expected agent")
	}
}

func TestVerify_NeverBootstrappedDirectory(t *testing.T) {
	target := t.TempDir()

	result, err := bootstrap.Verify(target, "fake-agent")
	if err != nil {
		t.Fatalf("Verify() unexpected error: %v", err)
	}
	if result.Detected {
		t.Error("Verify() Detected = true, want false for a directory that was never bootstrapped")
	}
}
