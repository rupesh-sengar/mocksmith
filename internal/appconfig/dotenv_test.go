package appconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDotEnvSetsValues(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := strings.Join([]string{
		"# comment",
		"ADMIN_KEY=from_file",
		"LISTEN_ADDR=:7777",
		"DATABASE_URL='postgres://u:p@localhost:5432/mocksmith?sslmode=disable'",
	}, "\n")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	unsetEnv(t, "ADMIN_KEY", "LISTEN_ADDR", "DATABASE_URL")

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("loadDotEnv returned error: %v", err)
	}

	if got := os.Getenv("ADMIN_KEY"); got != "from_file" {
		t.Fatalf("expected ADMIN_KEY from file, got %q", got)
	}
	if got := os.Getenv("LISTEN_ADDR"); got != ":7777" {
		t.Fatalf("expected LISTEN_ADDR from file, got %q", got)
	}
	if got := os.Getenv("DATABASE_URL"); got != "postgres://u:p@localhost:5432/mocksmith?sslmode=disable" {
		t.Fatalf("expected DATABASE_URL from file, got %q", got)
	}
}

func TestLoadDotEnvDoesNotOverrideExistingVars(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	content := "ADMIN_KEY=from_file"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := os.Setenv("ADMIN_KEY", "from_env"); err != nil {
		t.Fatalf("set ADMIN_KEY: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Unsetenv("ADMIN_KEY")
	})

	if err := loadDotEnv(path); err != nil {
		t.Fatalf("loadDotEnv returned error: %v", err)
	}

	if got := os.Getenv("ADMIN_KEY"); got != "from_env" {
		t.Fatalf("expected existing ADMIN_KEY to win, got %q", got)
	}
}

func TestLoadDotEnvMissingFileIsAllowed(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env-does-not-exist")
	if err := loadDotEnv(path); err != nil {
		t.Fatalf("expected nil on missing file, got: %v", err)
	}
}

func TestLoadDotEnvInvalidLine(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("NOT_A_VALID_LINE"), 0o600); err != nil {
		t.Fatalf("write .env: %v", err)
	}

	if err := loadDotEnv(path); err == nil {
		t.Fatalf("expected parse error for invalid line")
	}
}

func unsetEnv(t *testing.T, keys ...string) {
	t.Helper()

	type prev struct {
		value  string
		exists bool
	}
	previous := make(map[string]prev, len(keys))

	for _, key := range keys {
		v, ok := os.LookupEnv(key)
		previous[key] = prev{value: v, exists: ok}
		if err := os.Unsetenv(key); err != nil {
			t.Fatalf("unset %s: %v", key, err)
		}
	}

	t.Cleanup(func() {
		for _, key := range keys {
			p := previous[key]
			if !p.exists {
				_ = os.Unsetenv(key)
				continue
			}
			_ = os.Setenv(key, p.value)
		}
	})
}
