package bootstrap

// VerifyResult is the outcome of verifying a completed bootstrap: does
// the project detect successfully, and does the installed agent match
// what was requested.
type VerifyResult struct {
	// Detected is whether targetDir now detects as a valid project
	// (FR-010).
	Detected bool
	// InstalledAgent is the agent actually recorded as installed, per
	// Inspect.
	InstalledAgent string
	// AgentMatches is whether InstalledAgent equals the wantAgentID the
	// caller expected (FR-011, SC-005).
	AgentMatches bool
}

// Verify confirms a completed bootstrap is genuinely usable: targetDir
// detects as a valid project, and the agent actually installed matches
// wantAgentID (FR-011, SC-005).
//
// Verify is Inspect plus one comparison — not a second, independent
// detection or record-reading implementation (research.md).
func Verify(targetDir, wantAgentID string) (VerifyResult, error) {
	inspected, err := Inspect(targetDir)
	if err != nil {
		return VerifyResult{}, err
	}

	return VerifyResult{
		Detected:       inspected.Initialized,
		InstalledAgent: inspected.InstalledAgent,
		AgentMatches:   inspected.InstalledAgent == wantAgentID,
	}, nil
}
