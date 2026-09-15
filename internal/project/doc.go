// Package project implements misterspec's project-detection and
// configuration layer.
//
// It answers, deterministically, "is this directory inside an initialized
// misterspec project, where is its root, and what is its resolved
// configuration?" — the first question every other deterministic operation
// and every Skill needs answered before it can trust any other path.
//
// See specs/001-core-foundation/contracts/packages.md for the package's
// exported contract and specs/001-core-foundation/data-model.md for the
// Project/Configuration entity definitions.
package project
