package tui

import (
	"bytes"
	"errors"
	"io"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mottamarcio/misterspec/internal/agents"
	"github.com/mottamarcio/misterspec/internal/bootstrap"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/testutil"
)

// delayedKeyInput returns a reader that yields keys only after a short
// delay, then closes. ScreenInspect's own Update has no case for a
// keypress (Init()'s bootstrap.Inspect runs asynchronously, off the
// main loop) — a real terminal's human typing speed never collides with
// that sub-millisecond window, but a raw byte stream handed to the
// program at construction time can arrive before Init()'s result does,
// and would otherwise be silently dropped. The delay here is generous
// (well beyond Inspect's own local-filesystem cost) specifically so
// these tests never race it.
func delayedKeyInput(t *testing.T, keys string) io.Reader {
	t.Helper()
	r, w := io.Pipe()
	go func() {
		time.Sleep(50 * time.Millisecond)
		_, _ = io.WriteString(w, keys)
		w.Close()
	}()
	return r
}

// TestRunInit_HappyPathBootstrapsProject drives the full interactive
// flow via RunInit's own I/O-redirection options — no real terminal —
// against a fixture *agents.Registry and Skills fs.FS: select the one
// fixture agent (Enter), confirm the preview ("y"). The program quits
// on its own once bootstrap.Bootstrap's async result lands (Success
// auto-quits after rendering — contracts/tui.md's reconciliation), so
// no further synthetic input is needed. Confirms the target directory
// is genuinely bootstrapped afterward — quickstart.md's validation
// strategy, plan.md's T018.
func TestRunInit_HappyPathBootstrapsProject(t *testing.T) {
	target := t.TempDir()
	registry := agents.NewRegistry(fakeAdapter{id: "fake-agent", target: ".fake/skills"})

	var output bytes.Buffer

	err := RunInit(target, registry, fixtureSkills(),
		tea.WithInput(delayedKeyInput(t, "\r"+"y")),
		tea.WithOutput(&output),
		tea.WithoutSignalHandler(),
		tea.WithoutRenderer(),
	)
	if err != nil {
		t.Fatalf("RunInit() unexpected error: %v", err)
	}

	proj, err := project.Detect(target)
	if err != nil {
		t.Fatalf("project.Detect(target) after RunInit() unexpected error: %v", err)
	}
	if proj.Root != target {
		t.Errorf("proj.Root = %q, want %q", proj.Root, target)
	}
}

// TestRunInit_ContinueAnywayPastAlreadyInitializedReachesClearError
// completes spec.md's User Story 2 Acceptance Scenario 4: choosing
// "continue anyway" past the already-initialized Warning screen
// proceeds through agent selection and the preview, bootstrap.Bootstrap
// then rejects with its own already-proven ErrAlreadyInitialized
// (007-project-bootstrap), and the flow lands on ScreenError showing
// that specific, clear explanation — never a silent no-op or a forced
// overwrite (plan.md's Constitution Check, Principle VIII). Constructs
// the tea.Program directly (rather than via RunInit) so the test can
// inspect the final Model's own state, not just RunInit's bare error.
func TestRunInit_ContinueAnywayPastAlreadyInitializedReachesClearError(t *testing.T) {
	target := testutil.Project(t) // already an initialized project
	registry := agents.NewRegistry(fakeAdapter{id: "fake-agent", target: ".fake/skills"})

	// "c" (continue past the warning) + Enter (select the one fixture
	// agent) + "y" (confirm the preview) — the program then quits on
	// its own once Bootstrap's rejection lands.
	var output bytes.Buffer

	model := NewModel(target, registry, fixtureSkills())
	program := tea.NewProgram(model,
		tea.WithInput(delayedKeyInput(t, "c"+"\r"+"y")),
		tea.WithOutput(&output),
		tea.WithoutSignalHandler(),
		tea.WithoutRenderer(),
	)

	finalModel, err := program.Run()
	if err != nil {
		t.Fatalf("program.Run() unexpected error: %v", err)
	}

	final, ok := finalModel.(Model)
	if !ok {
		t.Fatalf("program.Run() returned %T, want Model", finalModel)
	}
	if final.screen != ScreenError {
		t.Fatalf("final screen = %v, want ScreenError", final.screen)
	}
	if !errors.Is(final.installErr, bootstrap.ErrAlreadyInitialized) {
		t.Errorf("installErr = %v, want errors.Is(_, bootstrap.ErrAlreadyInitialized)", final.installErr)
	}
}
