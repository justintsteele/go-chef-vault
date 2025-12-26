package vault

import (
	"testing"
)

func TestService_GetItem(t *testing.T) {
	setup()
	defer teardown()

	cleanup := stubVaultItemKeyDecrypt(t)
	defer cleanup()

	stubMuxGetItem(t)
	_, err := service.GetItem("vault1", "secret1")
	if err != nil {
		t.Fatalf("Vaults.GetItem returned error: %v", err)
	}
}
