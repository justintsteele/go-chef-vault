package cheferr

import (
	"errors"
	"net/http"

	"github.com/go-chef/chef"
)

// AsChefError attempts to extract a *chef.ErrorResponse from err.
//
// It first unwraps err using errors.As to support wrapped errors (%w),
// then falls back to go-chef's ChefError helper for compatibility.
// The returned boolean indicates whether the error originated from
// the Chef Server.
func AsChefError(err error) (*chef.ErrorResponse, bool) {
	if err == nil {
		return nil, false
	}

	var cerr *chef.ErrorResponse
	if errors.As(err, &cerr) {
		return cerr, true
	}

	if cerr, _ := chef.ChefError(err); cerr != nil {
		return cerr, true
	}

	return nil, false
}

// IsNotFound reports whether err represents a 404 response from the Chef Server.
func IsNotFound(err error) bool {
	if ce, ok := AsChefError(err); ok {
		return ce.Response.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict reports whether err represents a 409 conflict from the Chef Server.
func IsConflict(err error) bool {
	if ce, ok := AsChefError(err); ok {
		return ce.Response.StatusCode == http.StatusConflict
	}
	return false
}
