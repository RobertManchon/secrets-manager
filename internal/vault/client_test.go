// filepath: internal/vault/client_test.go

package vault

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClient(t *testing.T) {
	// Create a mock Vault server
	mockVault := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simply return 200 OK for connection test
		w.WriteHeader(http.StatusOK)
	}))
	defer mockVault.Close()

	tests := []struct {
		name        string
		config      *Config
		shouldError bool
	}{
		{
			name: "Valid configuration",
			config: &Config{
				Address: mockVault.URL,
				Token:   "test-token",
			},
			shouldError: false,
		},
		{
			name: "Empty address",
			config: &Config{
				Address: "",
				Token:   "test-token",
			},
			shouldError: true,
		},
		{
			name: "Empty token",
			config: &Config{
				Address: mockVault.URL,
				Token:   "",
			},
			shouldError: false, // Creating client with empty token should not fail
		},
		{
			name: "With namespace",
			config: &Config{
				Address:   mockVault.URL,
				Token:     "test-token",
				Namespace: "test-namespace",
			},
			shouldError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClient(tc.config)

			if tc.shouldError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				if client == nil {
					t.Error("Expected client to be returned, got nil")
				}
				if client != nil && client.client.Token() != tc.config.Token {
					t.Errorf("Expected token %s, got %s", tc.config.Token, client.client.Token())
				}
			}
		})
	}
}
