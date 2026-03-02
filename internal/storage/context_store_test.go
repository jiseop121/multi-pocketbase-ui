package storage

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestContextStoreSaveLoadAndClear(t *testing.T) {
	dir := t.TempDir()
	store := NewContextStore(dir)

	if err := store.Save(Context{DBAlias: "dev", SuperuserAlias: "root"}); err != nil {
		t.Fatalf("save: %v", err)
	}

	ctx, ok, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if !ok {
		t.Fatalf("expected context to exist")
	}
	if ctx.DBAlias != "dev" || ctx.SuperuserAlias != "root" {
		t.Fatalf("unexpected context: %+v", ctx)
	}
	if strings.TrimSpace(ctx.UpdatedAt) == "" {
		t.Fatalf("expected updatedAt to be populated")
	}

	if err := store.Clear(); err != nil {
		t.Fatalf("clear: %v", err)
	}
	_, ok, err = store.Load()
	if err != nil {
		t.Fatalf("load after clear: %v", err)
	}
	if ok {
		t.Fatalf("expected context to be removed")
	}
}

func TestContextStoreSaveRequiresDBAlias(t *testing.T) {
	store := NewContextStore(t.TempDir())
	if err := store.Save(Context{}); err == nil {
		t.Fatalf("expected validation error")
	}
}

func TestContextStoreLoadTreatsEmptyAsMissing(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "context.json")
	if err := os.WriteFile(path, []byte("\n"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	store := NewContextStore(dir)
	_, ok, err := store.Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if ok {
		t.Fatalf("expected no saved context")
	}
}
