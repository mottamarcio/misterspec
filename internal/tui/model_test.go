package tui

import (
	"testing"

	"github.com/mottamarcio/misterspec/internal/testutil"
)

func TestModelInit_ProducesInspectResultMsg(t *testing.T) {
	root := testutil.Project(t)
	m := NewModel(root, nil, nil)

	cmd := m.Init()
	if cmd == nil {
		t.Fatal("Init() returned a nil Cmd")
	}

	msg := cmd()
	result, ok := msg.(inspectResultMsg)
	if !ok {
		t.Fatalf("Init()'s Cmd produced %T, want inspectResultMsg", msg)
	}
	if result.err != nil {
		t.Fatalf("inspectResultMsg.err = %v, want nil", result.err)
	}
	if !result.result.Initialized {
		t.Error("inspectResultMsg.result.Initialized = false, want true for a fixture project")
	}
}

func TestModelInit_UninitializedDirectory(t *testing.T) {
	root := t.TempDir()
	m := NewModel(root, nil, nil)

	msg := m.Init()()
	result, ok := msg.(inspectResultMsg)
	if !ok {
		t.Fatalf("Init()'s Cmd produced %T, want inspectResultMsg", msg)
	}
	if result.result.Initialized {
		t.Error("inspectResultMsg.result.Initialized = true, want false")
	}
	if !result.result.Empty {
		t.Error("inspectResultMsg.result.Empty = false, want true")
	}
}
