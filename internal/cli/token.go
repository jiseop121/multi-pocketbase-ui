package cli

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"time"
)

func parseTokenExpiry(token string) (time.Time, bool) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return time.Time{}, false
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return time.Time{}, false
	}
	var claims map[string]any
	if err := json.Unmarshal(payload, &claims); err != nil {
		return time.Time{}, false
	}
	rawExp, ok := claims["exp"]
	if !ok {
		return time.Time{}, false
	}
	switch t := rawExp.(type) {
	case float64:
		if t <= 0 {
			return time.Time{}, false
		}
		return time.Unix(int64(t), 0).UTC(), true
	case int64:
		if t <= 0 {
			return time.Time{}, false
		}
		return time.Unix(t, 0).UTC(), true
	case int:
		if t <= 0 {
			return time.Time{}, false
		}
		return time.Unix(int64(t), 0).UTC(), true
	default:
		return time.Time{}, false
	}
}
