package operations

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/mottamarcio/misterspec/internal/artifacts"
	"github.com/mottamarcio/misterspec/internal/evidence"
	"github.com/mottamarcio/misterspec/internal/ids"
	"github.com/mottamarcio/misterspec/internal/project"
	"github.com/mottamarcio/misterspec/internal/vcs"
)

// defaultCaptureEvidenceTimeout is used when EvidenceCaptureRequest.Timeout
// is the zero value — contracts §4's own documented CLI default (2m).
const defaultCaptureEvidenceTimeout = 2 * time.Minute

// ErrInvalidEvidenceRequest is CaptureEvidence's own sentinel for a
// malformed EvidenceCaptureRequest — declared origin combined with
// Command, automated origin missing Command, declared origin missing
// Result, or a Task that does not exist in the named Spec's own
// tasks.md (041-task-evidence-fingerprint contracts §2).
var ErrInvalidEvidenceRequest = errors.New("operations: invalid evidence capture request")

// EvidenceCaptureRequest is CaptureEvidence's own input
// (041-task-evidence-fingerprint data-model.md "Evidence Capture
// Request/Result"). Exactly one of {Result} (Origin == "declared") or
// {Command, optionally Args/Dir/Timeout} (Origin == "automated")
// applies.
//
// SpecDir is not part of the original contract draft — added during
// implementation (correction): CaptureEvidence needs it to locate the
// Task's own current body in tasks.md and, once User Story 4 lands, to
// know where to write the evidence/ log directory. The CLI layer
// already computes this the same way internal prepare does
// (strings.TrimSuffix(loc.Path, "/spec.md")), so passing it explicitly
// avoids CaptureEvidence re-deriving it via a second Resolve call.
type EvidenceCaptureRequest struct {
	Spec    ids.EntityID
	Task    ids.EntityID
	SpecDir string
	Origin  string // "automated" | "declared"
	By      string
	Result  string // required iff Origin == "declared"
	Command string // required iff Origin == "automated"; forbidden otherwise
	Args    []string
	Dir     string
	Timeout time.Duration
}

// evidenceTaskHeadingPattern mirrors internal/prepare's and
// internal/validation's own (unexported) pattern of the same shape —
// duplicated here rather than imported, since internal/operations
// cannot import either without an import cycle (internal/prepare and
// internal/validation both sit above internal/operations in this
// codebase's own dependency graph: internal/context already imports
// internal/operations, and internal/prepare imports internal/context).
var evidenceTaskHeadingPattern = regexp.MustCompile(`^TASK-(\d+)\b`)

// currentTaskBody reads specDir's own tasks.md and returns task's own
// Section.Body — the exact same scope operations.TaskContentFingerprint
// hashes. Returns ErrEntityNotFound if the Task does not exist in this
// Spec's tasks.md.
func currentTaskBody(root, specDir string, task ids.EntityID) ([]byte, error) {
	tasksPath := filepath.Join(specDir, "tasks.md")
	body, err := artifacts.ReadBody(filepath.Join(root, tasksPath))
	if err != nil {
		return nil, err
	}
	doc := artifacts.ParseDocument(body)
	for _, section := range doc.Sections {
		sub := evidenceTaskHeadingPattern.FindStringSubmatch(section.Heading)
		if sub == nil {
			continue
		}
		n, convErr := strconv.Atoi(sub[1])
		if convErr != nil || n != task.Number {
			continue
		}
		return []byte(section.Body), nil
	}
	return nil, fmt.Errorf("%w: %v has no %v", ErrEntityNotFound, task, task)
}

// CaptureEvidence computes req.Task's current content fingerprint and
// Git snapshot, and — Origin == "automated" — runs Command with Args
// inside Dir (resolved and rejected if it would escape the project
// root, Principle VIII), bounded by Timeout, writing its combined
// output to a new <specDir>/evidence/<Task>-<timestamp>.log file.
// Returns the populated evidence.EvidenceFields. It never writes into
// tasks.md itself; that stays /mister-implement's own job
// (041-task-evidence-fingerprint research.md #4, contracts §2). A
// command that runs and exits non-zero, or that is killed for
// exceeding Timeout, still returns (fields, nil) with Result == "fail"
// — a successful capture of a failure, not an operational error.
func CaptureEvidence(root string, cfg project.Configuration, req EvidenceCaptureRequest) (evidence.EvidenceFields, error) {
	if req.Origin != "automated" && req.Origin != "declared" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: origin must be \"automated\" or \"declared\", got %q", ErrInvalidEvidenceRequest, req.Origin)
	}
	if req.By == "" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: by is required", ErrInvalidEvidenceRequest)
	}
	if req.Origin == "declared" && req.Command != "" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: declared origin must not set command", ErrInvalidEvidenceRequest)
	}
	if req.Origin == "declared" && req.Result != "pass" && req.Result != "fail" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: declared origin requires result \"pass\" or \"fail\", got %q", ErrInvalidEvidenceRequest, req.Result)
	}
	if req.Origin == "automated" && req.Command == "" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: automated origin requires command", ErrInvalidEvidenceRequest)
	}
	if req.Origin == "automated" && req.Result != "" {
		return evidence.EvidenceFields{}, fmt.Errorf("%w: automated origin must not set result — it is derived from the command's own exit code", ErrInvalidEvidenceRequest)
	}

	// Resolve and validate Dir before running anything at all — Principle
	// VIII: a path-traversal attempt must be rejected before any
	// subprocess ever starts, not discovered partway through.
	var execDirAbs string
	if req.Origin == "automated" {
		dirArg := req.Dir
		if dirArg == "" {
			dirArg = "."
		}
		rel, err := artifacts.RelativeWithinRoot(root, dirArg)
		if err != nil {
			return evidence.EvidenceFields{}, err
		}
		execDirAbs = filepath.Join(root, rel)
	}

	body, err := currentTaskBody(root, req.SpecDir, req.Task)
	if err != nil {
		return evidence.EvidenceFields{}, err
	}
	fp := TaskContentFingerprint(req.Task, body)

	gitRevision, available, err := vcs.HeadCommit(root)
	if err != nil {
		return evidence.EvidenceFields{}, err
	}
	workingTree := ""
	if available {
		dirty, err := vcs.IsWorkingTreeDirty(root)
		if err != nil {
			return evidence.EvidenceFields{}, err
		}
		if dirty {
			workingTree = "dirty"
		} else {
			workingTree = "clean"
		}
	}

	fields := evidence.EvidenceFields{
		Task:        req.Task,
		Result:      req.Result,
		Origin:      req.Origin,
		By:          req.By,
		CapturedAt:  time.Now().UTC().Format(time.RFC3339),
		GitRevision: gitRevision,
		WorkingTree: workingTree,
		Fingerprint: fp.String(),
	}

	if req.Origin == "automated" {
		result, command, logPath, err := runEvidenceCommand(root, req, execDirAbs)
		if err != nil {
			return evidence.EvidenceFields{}, err
		}
		fields.Result = result
		fields.Command = command
		fields.Log = logPath
	}

	return fields, nil
}

// runEvidenceCommand runs req.Command/req.Args inside execDirAbs,
// bounded by req.Timeout (defaultCaptureEvidenceTimeout if zero),
// captures its combined stdout+stderr, and writes that output to a new
// <specDir>/evidence/<Task>-<RFC3339-compact-timestamp>.log file via
// the existing writeAtomic helper (research.md #3, contracts §2).
// Returns "pass"/"fail" (never an error for a failing or timed-out
// command — only for an I/O failure writing the log).
func runEvidenceCommand(root string, req EvidenceCaptureRequest, execDirAbs string) (result, renderedCommand, logPath string, err error) {
	timeout := req.Timeout
	if timeout <= 0 {
		timeout = defaultCaptureEvidenceTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, req.Command, req.Args...)
	cmd.Dir = execDirAbs
	output, runErr := cmd.CombinedOutput()

	renderedCommand = strings.TrimSpace(req.Command + " " + strings.Join(req.Args, " "))
	if runErr != nil {
		result = "fail"
	} else {
		result = "pass"
	}

	timestamp := time.Now().UTC().Format("20060102T150405Z")
	logPath = filepath.ToSlash(filepath.Join(req.SpecDir, "evidence", fmt.Sprintf("%s-%s.log", req.Task, timestamp)))
	logAbs := filepath.Join(root, logPath)
	if err := writeAtomic(logAbs, string(output)); err != nil {
		return "", "", "", fmt.Errorf("operations: writing evidence log %s: %w", logPath, err)
	}

	return result, renderedCommand, logPath, nil
}
