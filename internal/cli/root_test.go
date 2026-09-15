package cli

import "testing"

func TestRootCmd_HelpListsOnlyPublicSurface(t *testing.T) {
	root := newRootCmd()

	var found []string
	for _, c := range root.Commands() {
		if !c.IsAvailableCommand() {
			continue // hidden, deprecated, or a bare help topic — not shown by --help
		}
		found = append(found, c.Name())
	}

	for _, name := range found {
		if name == "internal" {
			t.Errorf("root.Commands() (available/help-visible) includes %q, want it hidden", name)
		}
	}

	var hasInit bool
	for _, name := range found {
		if name == "init" {
			hasInit = true
		}
	}
	if !hasInit {
		t.Errorf("root.Commands() (available/help-visible) = %v, want it to include %q", found, "init")
	}
}

func TestRootCmd_InternalCommandStillRunsDirectly(t *testing.T) {
	// Hidden from help does not mean removed — invoking it directly by
	// name must still work (Acceptance Scenario 2).
	internalCmd, _, err := newRootCmd().Find([]string{"internal", "resolve"})
	if err != nil {
		t.Fatalf("Find([internal resolve]) unexpected error: %v", err)
	}
	if internalCmd == nil || internalCmd.Name() != "resolve" {
		t.Fatalf("Find([internal resolve]) = %v, want the resolve command", internalCmd)
	}
}
