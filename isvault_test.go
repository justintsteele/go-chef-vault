package vault

import "testing"

func TestService_IsVault(t *testing.T) {
	setupStubs(t)

	tests := []struct {
		name     string
		bag      string
		item     string
		expected bool
	}{
		{"vault item", "vault1", "secret1", true},
		{"encrypted data bag", "encrdata1", "encritem1", false},
		{"normal data bag", "databag1", "plaintext1", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := service.IsVault(tt.bag, tt.item)
			if err != nil {
				t.Fatal(err)
			}

			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
