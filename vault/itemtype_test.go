package vault

import "testing"

func TestService_ItemType(t *testing.T) {
	setupStubs(t)

	tests := []struct {
		name     string
		bag      string
		item     string
		expected DataBagItemType
	}{
		{"vault", "vault1", "secret1", DataBagItemTypeVault},
		{"encrypted data bag", "encrdata1", "encritem1", DataBagItemTypeEncrypted},
		{"normal data bag", "databag1", "plaintext1", DataBagItemTypeNormal},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.ItemType(tt.bag, tt.item)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
