// Package eval implements the deterministic half of the End-to-End
// Quality and Efficiency Evaluation harness (037-eval-quality-
// efficiency): loading retrieval EvaluationCase / agent-execution
// EvaluationTask definitions, running retrieval cases against the
// existing contextengine collector/ranker in-process, and comparing
// recorded RunRecords against a named Baseline.
//
// This package never launches, drives, or scrapes a live coding-agent
// session (Constitution Principle IV; spec037 research.md #1) — agent
// task-execution telemetry is always a RunRecord some external process
// already produced. eval only computes over already-recorded data.
package eval
