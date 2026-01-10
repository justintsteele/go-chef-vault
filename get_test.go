package vault

import (
	"testing"
)

func TestService_GetItem(t *testing.T) {
	setupStubs(t)

	_, err := service.GetItem("vault1", "secret1")
	if err != nil {
		t.Fatalf("Vaults.GetItem returned error: %v", err)
	}
}
