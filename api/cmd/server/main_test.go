package main

import "testing"

func TestToSolBotPostgresConfig(t *testing.T) {
	cfg, err := toSolBotPostgresConfig("postgres://user:pass@localhost:5432/sol_armada?sslmode=disable", 20)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if cfg.Host != "localhost" {
		t.Fatalf("Host = %q, expected %q", cfg.Host, "localhost")
	}
	if cfg.Port != 5432 {
		t.Fatalf("Port = %d, expected %d", cfg.Port, 5432)
	}
	if cfg.Username != "user" {
		t.Fatalf("Username = %q, expected %q", cfg.Username, "user")
	}
	if cfg.Password != "pass" {
		t.Fatalf("Password = %q, expected %q", cfg.Password, "pass")
	}
	if cfg.Database != "sol_armada" {
		t.Fatalf("Database = %q, expected %q", cfg.Database, "sol_armada")
	}
	if cfg.SSLMode != "disable" {
		t.Fatalf("SSLMode = %q, expected %q", cfg.SSLMode, "disable")
	}
	if cfg.MaxConns != 20 {
		t.Fatalf("MaxConns = %d, expected %d", cfg.MaxConns, 20)
	}
	if cfg.MinConns != 1 {
		t.Fatalf("MinConns = %d, expected %d", cfg.MinConns, 1)
	}
}

func TestToSolBotPostgresConfigErrors(t *testing.T) {
	tests := []struct {
		name string
		dsn  string
	}{
		{name: "invalid dsn", dsn: "::://bad"},
		{name: "missing host", dsn: "postgres://user:pass@:5432/db"},
		{name: "missing database", dsn: "postgres://user:pass@localhost:5432"},
		{name: "missing username", dsn: "postgres://:pass@localhost:5432/db"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toSolBotPostgresConfig(tt.dsn, 10)
			if err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}
