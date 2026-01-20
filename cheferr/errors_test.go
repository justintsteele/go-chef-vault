package cheferr

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/go-chef/chef"
)

func TestAsChefError_Wrapped(t *testing.T) {
	orig := &chef.ErrorResponse{
		Response: &http.Response{
			StatusCode: http.StatusConflict,
		},
	}

	err := fmt.Errorf("context: %w", orig)

	ce, ok := AsChefError(err)
	if !ok {
		t.Fatalf("expected wrapped error to be recognized as Chef error")
	}

	if !errors.Is(ce, orig) {
		t.Fatalf("expected original Chef error, got different instance")
	}
}
