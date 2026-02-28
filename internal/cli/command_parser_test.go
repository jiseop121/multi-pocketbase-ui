package cli

import (
	"reflect"
	"testing"
)

func TestParseCommandLinePreservesInnerQuotesInToken(t *testing.T) {
	line := "api records --filter status='open'"
	got, err := ParseCommandLine(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := []string{"api", "records", "--filter", "status='open'"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("tokens mismatch: got=%q want=%q", got, want)
	}
}

