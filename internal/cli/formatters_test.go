package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"multi-pocketbase-ui/internal/apperr"
	"multi-pocketbase-ui/internal/pocketbase"
)

func TestRenderQueryResultTableIncludesRows(t *testing.T) {
	result := pocketbase.QueryResult{Rows: []map[string]any{{"id": "1", "title": "hello"}}}
	out, err := RenderQueryResult("table", "", result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "1 rows") {
		t.Fatalf("missing rows suffix: %s", out)
	}
	if !strings.Contains(out, "ID") {
		t.Fatalf("missing table header: %s", out)
	}
}

func TestRenderQueryResultCSVRequiresOut(t *testing.T) {
	_, err := RenderQueryResult("csv", "", pocketbase.QueryResult{})
	if err == nil {
		t.Fatalf("expected error")
	}
	if apperr.ExitCode(err) != 2 {
		t.Fatalf("exit code mismatch: got=%d want=2", apperr.ExitCode(err))
	}
}

func TestRenderQueryResultEmptyTableShowsZeroRows(t *testing.T) {
	out, err := RenderQueryResult("table", "", pocketbase.QueryResult{Rows: []map[string]any{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "0 rows") {
		t.Fatalf("expected zero rows output: %s", out)
	}
}

func TestRenderQueryResultEmptyCSVUsesFallbackHeader(t *testing.T) {
	outPath := filepath.Join(t.TempDir(), "empty.csv")
	msg, err := RenderQueryResult("csv", outPath, pocketbase.QueryResult{Rows: []map[string]any{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(msg, "Exported 0 rows") {
		t.Fatalf("unexpected export summary: %s", msg)
	}
	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read csv: %v", err)
	}
	if !strings.HasPrefix(string(data), "result") {
		t.Fatalf("missing fallback header in csv: %q", string(data))
	}
}
