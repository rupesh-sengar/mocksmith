package appconfig

import "testing"

func TestConfigFromEnvDefaults(t *testing.T) {
	t.Parallel()

	cfg := configFromEnv(func(string) string { return "" })
	if cfg.AdminKey != "dev" {
		t.Fatalf("expected default admin key dev, got %q", cfg.AdminKey)
	}
	if cfg.ListenAddr != ":8787" {
		t.Fatalf("expected default listen addr :8787, got %q", cfg.ListenAddr)
	}
	if cfg.DatabaseURL != "" {
		t.Fatalf("expected empty database url, got %q", cfg.DatabaseURL)
	}
}

func TestConfigFromEnvValues(t *testing.T) {
	t.Parallel()

	env := map[string]string{
		"ADMIN_KEY":    "admin-123",
		"DATABASE_URL": "postgres://u:p@localhost:5432/mocksmith?sslmode=disable",
		"LISTEN_ADDR":  ":9999",
	}
	cfg := configFromEnv(func(k string) string { return env[k] })

	if cfg.AdminKey != "admin-123" {
		t.Fatalf("unexpected admin key: %q", cfg.AdminKey)
	}
	if cfg.DatabaseURL != env["DATABASE_URL"] {
		t.Fatalf("unexpected database url: %q", cfg.DatabaseURL)
	}
	if cfg.ListenAddr != ":9999" {
		t.Fatalf("unexpected listen addr: %q", cfg.ListenAddr)
	}
}
