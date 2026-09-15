---
id: LRN-001
type: learning
status: promoted
---

## Summary

Real correction discovered during 017-internal-context-command's own
implementation: a Cobra-level `IntVar` flag cannot be relied on to
report a non-numeric `--budget` value as a JSON error, because Cobra's
own automatic root-level flag-parse fallback only fires for a full
`root.Execute()` invocation — not for a leaf command's own `Execute()`,
which is how every existing per-command test (and, it turned out, the
real CLI test written for this exact command) actually exercises a
command. The `--budget` flag was corrected to a plain string, manually
parsed via `strconv.Atoi` inside `RunE`, returning the existing
`invalid_argument` JSON error directly — self-contained and testable
at the single-command level, with no dependency on the cross-cutting
root-level fallback.
