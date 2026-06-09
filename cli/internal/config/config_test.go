package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeConfigFile(t *testing.T, tmpHome string, cfg *Config) {
	t.Helper()
	dir := filepath.Join(tmpHome, ".enparse")
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("json.MarshalIndent: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"), data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
}

func clearEnvOverrides(t *testing.T) {
	t.Helper()
	for _, v := range []string{"ENPARSE_RPC_URL", "ENPARSE_REGISTRY_ADDR", "ENPARSE_VAULT_ADDR"} {
		t.Setenv(v, "")
	}
}

// ─── Load — file missing ─────────────────────────────────────────────────────

func TestLoadNoFileDefaultsRPCURL(t *testing.T) {
	clearEnvOverrides(t)
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RPCURL != "http://127.0.0.1:8545" {
		t.Errorf("RPCURL = %q, want default http://127.0.0.1:8545", cfg.RPCURL)
	}
}

func TestLoadNoFileAddressesEmpty(t *testing.T) {
	clearEnvOverrides(t)
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.IdentityRegistryAddr != "" || cfg.ProjectVaultAddr != "" {
		t.Error("expected all address fields empty when no config file")
	}
}

// ─── Load — file present ─────────────────────────────────────────────────────

func TestLoadReadsFileValues(t *testing.T) {
	clearEnvOverrides(t)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	want := &Config{
		RPCURL:               "https://sepolia.infura.io/v3/abc",
		IdentityRegistryAddr: "0xBB",
		ProjectVaultAddr:     "0xCC",
	}
	writeConfigFile(t, tmpHome, want)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RPCURL != want.RPCURL {
		t.Errorf("RPCURL = %q, want %q", cfg.RPCURL, want.RPCURL)
	}
	if cfg.IdentityRegistryAddr != want.IdentityRegistryAddr {
		t.Errorf("IdentityRegistryAddr = %q, want %q", cfg.IdentityRegistryAddr, want.IdentityRegistryAddr)
	}
	if cfg.ProjectVaultAddr != want.ProjectVaultAddr {
		t.Errorf("ProjectVaultAddr = %q, want %q", cfg.ProjectVaultAddr, want.ProjectVaultAddr)
	}
}

// ─── Env var overrides ───────────────────────────────────────────────────────

func TestEnvRPCURLOverridesFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	writeConfigFile(t, tmpHome, &Config{RPCURL: "http://from-file"})

	t.Setenv("ENPARSE_RPC_URL", "https://sepolia.example.com")
	t.Setenv("ENPARSE_REGISTRY_ADDR", "")
	t.Setenv("ENPARSE_VAULT_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RPCURL != "https://sepolia.example.com" {
		t.Errorf("RPCURL = %q, want env override", cfg.RPCURL)
	}
}

func TestEnvRegistryAddrOverridesFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	writeConfigFile(t, tmpHome, &Config{IdentityRegistryAddr: "0xFILE"})

	t.Setenv("ENPARSE_RPC_URL", "")
	t.Setenv("ENPARSE_REGISTRY_ADDR", "0xENV_REG")
	t.Setenv("ENPARSE_VAULT_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.IdentityRegistryAddr != "0xENV_REG" {
		t.Errorf("IdentityRegistryAddr = %q, want 0xENV_REG", cfg.IdentityRegistryAddr)
	}
}

func TestEnvVaultAddrOverridesFile(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	writeConfigFile(t, tmpHome, &Config{ProjectVaultAddr: "0xFILE"})

	t.Setenv("ENPARSE_RPC_URL", "")
	t.Setenv("ENPARSE_REGISTRY_ADDR", "")
	t.Setenv("ENPARSE_VAULT_ADDR", "0xENV_VAULT")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.ProjectVaultAddr != "0xENV_VAULT" {
		t.Errorf("ProjectVaultAddr = %q, want 0xENV_VAULT", cfg.ProjectVaultAddr)
	}
}

func TestEmptyEnvDoesNotOverrideFileValue(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)
	writeConfigFile(t, tmpHome, &Config{RPCURL: "https://from-file"})

	t.Setenv("ENPARSE_RPC_URL", "")
	t.Setenv("ENPARSE_REGISTRY_ADDR", "")
	t.Setenv("ENPARSE_VAULT_ADDR", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.RPCURL != "https://from-file" {
		t.Errorf("RPCURL = %q, want file value preserved when env is empty", cfg.RPCURL)
	}
}

// ─── Validate ────────────────────────────────────────────────────────────────

func TestValidatePassesWhenAllFieldsSet(t *testing.T) {
	cfg := &Config{
		RPCURL:               "https://rpc.example.com",
		IdentityRegistryAddr: "0xAA",
		ProjectVaultAddr:     "0xBB",
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("Validate returned unexpected error: %v", err)
	}
}

func TestValidateRejectsEmptyRPCURL(t *testing.T) {
	cfg := &Config{
		IdentityRegistryAddr: "0xAA",
		ProjectVaultAddr:     "0xBB",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty rpc_url, got nil")
	}
}

func TestValidateRejectsEmptyRegistryAddr(t *testing.T) {
	cfg := &Config{
		RPCURL:           "https://rpc.example.com",
		ProjectVaultAddr: "0xBB",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty identity_registry_addr, got nil")
	}
}

func TestValidateRejectsEmptyVaultAddr(t *testing.T) {
	cfg := &Config{
		RPCURL:               "https://rpc.example.com",
		IdentityRegistryAddr: "0xAA",
	}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected error for empty project_vault_addr, got nil")
	}
}

// ─── Save ────────────────────────────────────────────────────────────────────

func TestSaveWritesReadableJSON(t *testing.T) {
	clearEnvOverrides(t)
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	original := &Config{
		RPCURL:               "https://save-test",
		IdentityRegistryAddr: "0xF2",
		ProjectVaultAddr:     "0xF3",
	}
	if err := original.Save(); err != nil {
		t.Fatalf("Save: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(tmpHome, ".enparse", "config.json"))
	if err != nil {
		t.Fatalf("ReadFile after Save: %v", err)
	}

	var loaded Config
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}

	if loaded.RPCURL != original.RPCURL {
		t.Errorf("RPCURL = %q, want %q", loaded.RPCURL, original.RPCURL)
	}
	if loaded.IdentityRegistryAddr != original.IdentityRegistryAddr {
		t.Errorf("IdentityRegistryAddr mismatch after Save")
	}
	if loaded.ProjectVaultAddr != original.ProjectVaultAddr {
		t.Errorf("ProjectVaultAddr mismatch after Save")
	}
}
