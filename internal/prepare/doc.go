// Package prepare assembles everything an agent needs to start
// executing one Task in a single, read-only response
// (034-task-oriented-context-preparation) — Task identity, readiness
// (blocked/ready with named blockers), served requirements' own text,
// declared scope and verification method, and Plan sections associated
// with the same requirements. It orchestrates internal/operations,
// internal/validation, and internal/context's own primitives; it owns
// no new persisted state (Constitution Principle III) and executes no
// command of its own (spec FR-002).
package prepare
