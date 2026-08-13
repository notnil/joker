// Package table provides a modular, testable poker table engine.
//
// Goals:
//   - Simple seating: seat, autoseat, leave, occupancy queries
//   - Deterministic action flow: dealer rotation, next to act, start/end hand
//   - Clear state snapshots for read-only consumption
package table

