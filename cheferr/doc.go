// Package cheferr provides helpers for interpreting errors returned by the
// go-chef client in a consistent, semantic way.
//
// go-chef exposes Chef server errors as *chef.ErrorResponse, but its helper
// only performs a direct type assertion and does not unwrap wrapped errors.
// This package bridges that gap so callers can reason about Chef failures
// (e.g. NotFound, Conflict) without duplicating error inspection logic.
package cheferr
