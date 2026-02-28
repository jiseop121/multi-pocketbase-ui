package pocketbase

import (
	"strings"
	"testing"
)

func TestBuildEndpointsEscapeDynamicSegments(t *testing.T) {
	collectionEndpoint := BuildCollectionEndpoint("a/b")
	if strings.Contains(collectionEndpoint, "/a/b") {
		t.Fatalf("collection endpoint should escape slash: %s", collectionEndpoint)
	}
	if !strings.Contains(collectionEndpoint, "a%2Fb") {
		t.Fatalf("collection endpoint missing escaped segment: %s", collectionEndpoint)
	}

	recordsEndpoint := BuildRecordsEndpoint("x y")
	if !strings.Contains(recordsEndpoint, "x%20y") {
		t.Fatalf("records endpoint missing escaped segment: %s", recordsEndpoint)
	}

	recordEndpoint := BuildRecordEndpoint("../posts", "rec/1")
	if strings.Contains(recordEndpoint, "../") || strings.Contains(recordEndpoint, "/rec/1") {
		t.Fatalf("record endpoint should escape traversal and slash: %s", recordEndpoint)
	}
	if !strings.Contains(recordEndpoint, "..%2Fposts") {
		t.Fatalf("record endpoint should contain escaped traversal: %s", recordEndpoint)
	}
	if !strings.Contains(recordEndpoint, "rec%2F1") {
		t.Fatalf("record endpoint should contain escaped id: %s", recordEndpoint)
	}
}

func TestCollectColumnsPriorityOrdering(t *testing.T) {
	rows := []map[string]any{{
		"z":       "v",
		"created": "2026-01-01",
		"title":   "hello",
		"id":      "1",
		"a":       "b",
	}}

	cols := CollectColumns(rows)
	want := []string{"id", "title", "created", "a", "z"}
	if len(cols) != len(want) {
		t.Fatalf("column count mismatch: got=%d want=%d (%v)", len(cols), len(want), cols)
	}
	for i := range want {
		if cols[i] != want[i] {
			t.Fatalf("column order mismatch at %d: got=%q want=%q (%v)", i, cols[i], want[i], cols)
		}
	}
}
