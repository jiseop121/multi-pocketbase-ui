package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSuperuserStoreEncryptsPasswordAtRest(t *testing.T) {
	dir := t.TempDir()
	store := NewSuperuserStore(dir)

	if err := store.Add("dev", "root", "root@example.com", "secret-password"); err != nil {
		t.Fatalf("add superuser: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "superusers.json"))
	if err != nil {
		t.Fatalf("read superusers.json: %v", err)
	}
	text := string(raw)
	if strings.Contains(text, "secret-password") {
		t.Fatalf("password should not be stored in plaintext: %s", text)
	}
	if strings.Contains(text, `"password"`) {
		t.Fatalf("legacy password field should not be persisted: %s", text)
	}
	if !strings.Contains(text, `"passwordEnc"`) {
		t.Fatalf("encrypted password field missing: %s", text)
	}

	found, ok, err := store.Find("dev", "root")
	if err != nil {
		t.Fatalf("find superuser: %v", err)
	}
	if !ok {
		t.Fatalf("expected superuser to be found")
	}
	if found.Password != "secret-password" {
		t.Fatalf("decrypted password mismatch: got=%q", found.Password)
	}
}

func TestSuperuserStoreMigratesLegacyPlaintextOnWrite(t *testing.T) {
	dir := t.TempDir()
	legacy := `[
  {
    "dbAlias": "dev",
    "alias": "legacy",
    "email": "legacy@example.com",
    "password": "legacy-password"
  }
]
`
	if err := os.WriteFile(filepath.Join(dir, "superusers.json"), []byte(legacy), 0o600); err != nil {
		t.Fatalf("write legacy superusers.json: %v", err)
	}

	store := NewSuperuserStore(dir)
	if err := store.Add("dev", "new", "new@example.com", "new-password"); err != nil {
		t.Fatalf("add superuser with legacy file: %v", err)
	}

	raw, err := os.ReadFile(filepath.Join(dir, "superusers.json"))
	if err != nil {
		t.Fatalf("read superusers.json: %v", err)
	}
	text := string(raw)
	if strings.Contains(text, "legacy-password") || strings.Contains(text, "new-password") {
		t.Fatalf("plaintext password should be migrated out: %s", text)
	}
	if strings.Contains(text, `"password"`) {
		t.Fatalf("legacy password field should be removed after write: %s", text)
	}
}
