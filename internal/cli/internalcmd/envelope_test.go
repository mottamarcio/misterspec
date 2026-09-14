package internalcmd_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/mottamarcio/misterspec/internal/cli/internalcmd"
	"github.com/mottamarcio/misterspec/internal/operations"
)

func TestWriteSuccess_Envelope(t *testing.T) {
	var buf bytes.Buffer

	err := internalcmd.WriteSuccess(&buf, map[string]any{
		"entity": map[string]any{"id": "SPEC-014", "type": "spec", "path": "ai/.../spec.md"},
	})
	if err != nil {
		t.Fatalf("WriteSuccess() unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v (got %q)", err, buf.String())
	}
	if decoded["ok"] != true {
		t.Errorf(`decoded["ok"] = %v, want true`, decoded["ok"])
	}
	entity, ok := decoded["entity"].(map[string]any)
	if !ok {
		t.Fatalf(`decoded["entity"] = %v, want an object`, decoded["entity"])
	}
	if entity["id"] != "SPEC-014" {
		t.Errorf(`entity["id"] = %v, want "SPEC-014"`, entity["id"])
	}
}

func TestWriteSuccess_MultipleFields(t *testing.T) {
	var buf bytes.Buffer

	err := internalcmd.WriteSuccess(&buf, map[string]any{
		"valid":    false,
		"findings": []string{},
	})
	if err != nil {
		t.Fatalf("WriteSuccess() unexpected error: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(buf.Bytes(), &decoded); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if decoded["ok"] != true {
		t.Errorf(`decoded["ok"] = %v, want true`, decoded["ok"])
	}
	if decoded["valid"] != false {
		t.Errorf(`decoded["valid"] = %v, want false`, decoded["valid"])
	}
	if _, present := decoded["findings"]; !present {
		t.Error(`decoded["findings"] missing, want present`)
	}
}

func TestWriteError_Envelope(t *testing.T) {
	var buf bytes.Buffer

	err := internalcmd.WriteError(&buf, operations.ErrEntityNotFound)

	var decoded map[string]any
	if jsonErr := json.Unmarshal(buf.Bytes(), &decoded); jsonErr != nil {
		t.Fatalf("output is not valid JSON: %v (got %q)", jsonErr, buf.String())
	}
	if decoded["ok"] != false {
		t.Errorf(`decoded["ok"] = %v, want false`, decoded["ok"])
	}
	errObj, ok := decoded["error"].(map[string]any)
	if !ok {
		t.Fatalf(`decoded["error"] = %v, want an object`, decoded["error"])
	}
	if errObj["code"] != "entity_not_found" {
		t.Errorf(`error.code = %v, want "entity_not_found"`, errObj["code"])
	}
	if errObj["message"] == "" || errObj["message"] == nil {
		t.Error("error.message is empty, want a concise explanation")
	}

	var exitErr *internalcmd.ExitCodeError
	if !errors.As(err, &exitErr) {
		t.Fatalf("WriteError() returned error = %v, want an *ExitCodeError", err)
	}
	if exitErr.Code != 3 {
		t.Errorf("ExitCodeError.Code = %d, want 3 (entity_not_found)", exitErr.Code)
	}
}
