package pocketbase

import "testing"

func TestJoinURLPreservesBasePathPrefix(t *testing.T) {
	u, err := joinURL("https://example.com/pb", "/api/collections")
	if err != nil {
		t.Fatalf("joinURL error: %v", err)
	}
	if u != "https://example.com/pb/api/collections" {
		t.Fatalf("unexpected url: %s", u)
	}
}
