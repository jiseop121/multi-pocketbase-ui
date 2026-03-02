package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Context struct {
	DBAlias        string `json:"dbAlias"`
	SuperuserAlias string `json:"superuserAlias,omitempty"`
	UpdatedAt      string `json:"updatedAt"`
}

type ContextStore struct {
	path string
	now  func() time.Time
}

func NewContextStore(dataDir string) *ContextStore {
	return &ContextStore{
		path: filepath.Join(dataDir, "context.json"),
		now:  time.Now,
	}
}

func (s *ContextStore) Load() (Context, bool, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return Context{}, false, nil
		}
		return Context{}, false, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return Context{}, false, nil
	}

	var saved Context
	if err := json.Unmarshal(data, &saved); err != nil {
		return Context{}, false, err
	}
	if strings.TrimSpace(saved.DBAlias) == "" {
		return Context{}, false, nil
	}
	return saved, true, nil
}

func (s *ContextStore) Save(ctx Context) error {
	if strings.TrimSpace(ctx.DBAlias) == "" {
		return NewValidationError("context db alias is required")
	}
	if strings.TrimSpace(ctx.UpdatedAt) == "" {
		ctx.UpdatedAt = s.now().UTC().Format(time.RFC3339)
	}

	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, append(data, '\n'), 0o600)
}

func (s *ContextStore) Clear() error {
	err := os.Remove(s.path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
