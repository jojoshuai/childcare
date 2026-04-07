package config

import (
	"os"
	"testing"
)

// setEnv sets a map of env vars and returns a cleanup function that restores
// the original values.
func setEnv(t *testing.T, vars map[string]string) func() {
	t.Helper()
	originals := make(map[string]string, len(vars))
	for k := range vars {
		originals[k] = os.Getenv(k)
	}
	for k, v := range vars {
		if v == "" {
			os.Unsetenv(k)
		} else {
			os.Setenv(k, v)
		}
	}
	return func() {
		for k, v := range originals {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}
}

// allRequired returns a complete set of required env vars for use in tests.
func allRequired() map[string]string {
	return map[string]string{
		"MYSQL_DSN":          "root:pass@tcp(localhost:3306)/test?parseTime=true",
		"JWT_SECRET":         "test-jwt-secret",
		"JWT_REFRESH_SECRET": "test-refresh-secret",
		"WX_APPID":           "test-appid",
		"WX_SECRET":          "test-wx-secret",
		"PORT":               "9090",
	}
}

func TestLoad_AllVarsPresent(t *testing.T) {
	cleanup := setEnv(t, allRequired())
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.MYSQDSN != "root:pass@tcp(localhost:3306)/test?parseTime=true" {
		t.Errorf("unexpected MYSQL_DSN: %q", cfg.MYSQDSN)
	}
	if cfg.JWTSecret != "test-jwt-secret" {
		t.Errorf("unexpected JWT_SECRET: %q", cfg.JWTSecret)
	}
	if cfg.JWTRefreshSecret != "test-refresh-secret" {
		t.Errorf("unexpected JWT_REFRESH_SECRET: %q", cfg.JWTRefreshSecret)
	}
	if cfg.WXAppID != "test-appid" {
		t.Errorf("unexpected WX_APPID: %q", cfg.WXAppID)
	}
	if cfg.WXSecret != "test-wx-secret" {
		t.Errorf("unexpected WX_SECRET: %q", cfg.WXSecret)
	}
	if cfg.Port != "9090" {
		t.Errorf("unexpected PORT: %q", cfg.Port)
	}
}

func TestLoad_MissingMySQLDSN(t *testing.T) {
	vars := allRequired()
	vars["MYSQL_DSN"] = ""
	cleanup := setEnv(t, vars)
	defer cleanup()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing MYSQL_DSN, got nil")
	}
}

func TestLoad_MissingJWTSecret(t *testing.T) {
	vars := allRequired()
	vars["JWT_SECRET"] = ""
	cleanup := setEnv(t, vars)
	defer cleanup()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_SECRET, got nil")
	}
}

func TestLoad_MissingJWTRefreshSecret(t *testing.T) {
	vars := allRequired()
	vars["JWT_REFRESH_SECRET"] = ""
	cleanup := setEnv(t, vars)
	defer cleanup()

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing JWT_REFRESH_SECRET, got nil")
	}
}

func TestLoad_PortDefaultsTo8080(t *testing.T) {
	vars := allRequired()
	vars["PORT"] = ""
	cleanup := setEnv(t, vars)
	defer cleanup()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("expected PORT to default to \"8080\", got %q", cfg.Port)
	}
}

func TestLoad_OptionalWXVarsMissing(t *testing.T) {
	vars := allRequired()
	vars["WX_APPID"] = ""
	vars["WX_SECRET"] = ""
	cleanup := setEnv(t, vars)
	defer cleanup()

	// Should succeed — WX vars are optional
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error when WX vars are absent, got: %v", err)
	}
	if cfg.WXAppID != "" {
		t.Errorf("expected empty WX_APPID, got %q", cfg.WXAppID)
	}
	if cfg.WXSecret != "" {
		t.Errorf("expected empty WX_SECRET, got %q", cfg.WXSecret)
	}
}
