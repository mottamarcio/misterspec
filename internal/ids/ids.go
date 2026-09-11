package ids

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// ErrInvalidIDSyntax is returned by Parse when raw does not match the
// fixed prefix, all-numeric suffix, and exact zero-padding width required
// for entity type t (FR-010) — a wrong prefix, a non-numeric suffix, or a
// suffix of the wrong length are all rejected rather than accepted.
var ErrInvalidIDSyntax = errors.New("ids: invalid ID syntax")

// Parse validates raw against the syntax required for entity type t (its
// fixed prefix, a "-", and exactly width numeric digits) and returns the
// resulting EntityID. It never returns a partially valid EntityID: any
// syntax problem is reported as ErrInvalidIDSyntax instead.
func Parse(t EntityType, raw string, width int) (EntityID, error) {
	prefix := t.Prefix()
	if prefix == "" {
		return EntityID{}, fmt.Errorf("ids: unsupported entity type %v", t)
	}

	wantPrefix := prefix + "-"
	if !strings.HasPrefix(raw, wantPrefix) {
		return EntityID{}, fmt.Errorf("%w: %q does not start with %q", ErrInvalidIDSyntax, raw, wantPrefix)
	}

	suffix := raw[len(wantPrefix):]
	if len(suffix) != width {
		return EntityID{}, fmt.Errorf("%w: %q suffix has width %d, want %d", ErrInvalidIDSyntax, raw, len(suffix), width)
	}
	for _, r := range suffix {
		if r < '0' || r > '9' {
			return EntityID{}, fmt.Errorf("%w: %q suffix is not all-numeric", ErrInvalidIDSyntax, raw)
		}
	}

	number, err := strconv.Atoi(suffix)
	if err != nil {
		return EntityID{}, fmt.Errorf("%w: %q: %v", ErrInvalidIDSyntax, raw, err)
	}
	if number < 1 {
		return EntityID{}, fmt.Errorf("%w: %q number must be >= 1", ErrInvalidIDSyntax, raw)
	}

	return EntityID{Type: t, Prefix: prefix, Number: number, Width: width}, nil
}

// ParseAny parses raw (e.g. "SPEC-014"), deriving both its EntityType
// (from the prefix, via TypeForPrefix) and its width (from the suffix
// length) from the string itself — unlike Parse, the caller does not need
// to already know raw's type or width. This is what any consumer that
// only has a bare ID string in hand (an agent, a future CLI argument)
// needs; Parse remains for a caller that already knows the exact type and
// width it expects and wants that assumption enforced.
func ParseAny(raw string) (EntityID, error) {
	idx := strings.LastIndex(raw, "-")
	if idx <= 0 || idx == len(raw)-1 {
		return EntityID{}, fmt.Errorf("%w: %q is not a well-formed entity ID", ErrInvalidIDSyntax, raw)
	}
	prefix, suffix := raw[:idx], raw[idx+1:]

	t, ok := TypeForPrefix(prefix)
	if !ok {
		return EntityID{}, fmt.Errorf("%w: %q has an unrecognized entity prefix %q", ErrInvalidIDSyntax, raw, prefix)
	}

	return Parse(t, raw, len(suffix))
}

// NextID computes the next available ID for entity type t as a pure
// function of existing: the highest Number among existing entries of type
// t, plus one — or 1 if none exist. It never reads or writes any
// persisted counter (Constitution Principle III, FR-013).
func NextID(existing []EntityID, t EntityType, width int) EntityID {
	max := 0
	for _, id := range existing {
		if id.Type != t {
			continue
		}
		if id.Number > max {
			max = id.Number
		}
	}
	return EntityID{Type: t, Prefix: t.Prefix(), Number: max + 1, Width: width}
}
