package handlers

import "testing"

func TestHasRole(t *testing.T) {
	tests := []struct {
		name     string
		roles    []string
		role     string
		expected bool
	}{
		{
			name:     "finds existing role",
			roles:    []string{"member", "admin"},
			role:     "admin",
			expected: true,
		},
		{
			name:     "returns false when role missing",
			roles:    []string{"member", "moderator"},
			role:     "admin",
			expected: false,
		},
		{
			name:     "returns false for empty roles",
			roles:    []string{},
			role:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasRole(tt.roles, tt.role)
			if got != tt.expected {
				t.Fatalf("hasRole() = %v, expected %v", got, tt.expected)
			}
		})
	}
}
