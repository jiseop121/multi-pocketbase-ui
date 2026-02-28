package cli

import (
	"fmt"
	"strings"
	"unicode"
)

func ParseCommandLine(line string) ([]string, error) {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return nil, nil
	}

	var (
		tokens       []string
		current      []rune
		inSingle     bool
		inDouble     bool
		escaped      bool
		quoteStarted bool
		keepSingle   bool
		keepDouble   bool
	)

	flush := func(force bool) {
		if len(current) > 0 || (force && quoteStarted) {
			tokens = append(tokens, string(current))
			current = current[:0]
			quoteStarted = false
		}
	}

	for _, r := range trimmed {
		switch {
		case escaped:
			current = append(current, r)
			escaped = false
		case r == '\\' && !inSingle:
			escaped = true
		case r == '\'' && !inDouble:
			if inSingle {
				if keepSingle {
					current = append(current, r)
				}
				inSingle = false
				keepSingle = false
				continue
			}
			inSingle = true
			quoteStarted = true
			if len(current) > 0 {
				keepSingle = true
				current = append(current, r)
			}
		case r == '"' && !inSingle:
			if inDouble {
				if keepDouble {
					current = append(current, r)
				}
				inDouble = false
				keepDouble = false
				continue
			}
			inDouble = true
			quoteStarted = true
			if len(current) > 0 {
				keepDouble = true
				current = append(current, r)
			}
		case unicode.IsSpace(r) && !inSingle && !inDouble:
			flush(false)
		default:
			current = append(current, r)
		}
	}

	if escaped || inSingle || inDouble {
		return nil, fmt.Errorf("unterminated quote or escape sequence")
	}
	flush(false)
	return tokens, nil
}
