package cli

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"
)

func TestParseTokenExpiry(t *testing.T) {
	exp := time.Date(2026, 3, 2, 12, 0, 0, 0, time.UTC)
	token := makeJWT(t, exp.Unix())

	got, ok := parseTokenExpiry(token)
	if !ok {
		t.Fatalf("expected expiry")
	}
	if !got.Equal(exp) {
		t.Fatalf("expiry mismatch: got=%s want=%s", got, exp)
	}
}

func TestParseTokenExpiryReturnsFalseWhenMissingExp(t *testing.T) {
	token := makeJWTRawPayload(t, `{"sub":"user"}`)
	if _, ok := parseTokenExpiry(token); ok {
		t.Fatalf("expected no expiry")
	}
}

func TestParseTokenExpiryReturnsFalseOnInvalidToken(t *testing.T) {
	if _, ok := parseTokenExpiry("not-a-jwt"); ok {
		t.Fatalf("expected no expiry")
	}
}

func makeJWT(t *testing.T, exp int64) string {
	t.Helper()
	return makeJWTRawPayload(t, fmt.Sprintf(`{"exp":%d}`, exp))
}

func makeJWTRawPayload(t *testing.T, payload string) string {
	t.Helper()
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"none","typ":"JWT"}`))
	body := base64.RawURLEncoding.EncodeToString([]byte(payload))
	return header + "." + body + ".sig"
}
