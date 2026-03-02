package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"multi-pocketbase-ui/internal/apperr"
	"multi-pocketbase-ui/internal/pocketbase"
	"multi-pocketbase-ui/internal/storage"
)

type DispatcherConfig struct {
	Stdout  io.Writer
	Version string
	DataDir string
}

type commandContext struct {
	DBAlias        string
	SuperuserAlias string
}

type authCacheKey struct {
	dbAlias string
	suAlias string
}

type authCacheEntry struct {
	token     string
	expiresAt time.Time
	hasExpiry bool
}

type Dispatcher struct {
	stdout   io.Writer
	version  string
	dbStore  *storage.DBStore
	suStore  *storage.SuperuserStore
	ctxStore *storage.ContextStore
	pbClient *pocketbase.Client

	sessionCtx commandContext
	savedCtx   commandContext
	hasSaved   bool

	isREPL bool
	isTTY  bool

	authCache map[authCacheKey]authCacheEntry
	now       func() time.Time
}

func NewDispatcher(cfg DispatcherConfig) *Dispatcher {
	d := &Dispatcher{
		stdout:    cfg.Stdout,
		version:   cfg.Version,
		dbStore:   storage.NewDBStore(cfg.DataDir),
		suStore:   storage.NewSuperuserStore(cfg.DataDir),
		ctxStore:  storage.NewContextStore(cfg.DataDir),
		pbClient:  pocketbase.NewClient(),
		authCache: map[authCacheKey]authCacheEntry{},
		now:       time.Now,
	}
	if saved, ok, err := d.ctxStore.Load(); err == nil && ok {
		d.savedCtx = commandContext{DBAlias: saved.DBAlias, SuperuserAlias: saved.SuperuserAlias}
		d.hasSaved = true
	}
	return d
}

func (d *Dispatcher) SetREPLRuntime(isREPL, isTTY bool) {
	d.isREPL = isREPL
	d.isTTY = isTTY
}

func (d *Dispatcher) IsInteractiveTTY() bool {
	return d.isREPL && d.isTTY
}

func (d *Dispatcher) Execute(ctx context.Context, line string) error {
	tokens, err := ParseCommandLine(line)
	if err != nil {
		return apperr.Invalid("Could not parse command line.", "Check quotes and escape characters.")
	}
	if len(tokens) == 0 {
		return nil
	}
	if tokens[0] == "pbviewer" {
		tokens = tokens[1:]
		if len(tokens) == 0 {
			return nil
		}
	}

	switch tokens[0] {
	case "help":
		d.printHelp()
		return nil
	case "version":
		_, _ = fmt.Fprintln(d.stdout, d.version)
		return nil
	case "ui":
		return apperr.Invalid("UI mode is not available in Track 1.", "")
	case "db":
		return d.execDB(argsAfterHead(tokens))
	case "superuser":
		return d.execSuperuser(argsAfterHead(tokens))
	case "context":
		return d.execContext(argsAfterHead(tokens))
	case "api":
		return d.execAPI(ctx, argsAfterHead(tokens))
	case "exit", "quit":
		return ErrExitRequested
	default:
		return apperr.Invalid("Unknown command `"+tokens[0]+"`.", "Run `help` to see available commands.")
	}
}

func argsAfterHead(tokens []string) []string {
	if len(tokens) <= 1 {
		return []string{}
	}
	return tokens[1:]
}

func (d *Dispatcher) execDB(args []string) error {
	if len(args) == 0 {
		return apperr.Invalid("Missing db subcommand.", "Use: db add|list|remove")
	}
	cmd := args[0]
	switch cmd {
	case "add":
		fs := newFlagSet("db add")
		alias := fs.String("alias", "", "")
		baseURL := fs.String("url", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *alias == "" || *baseURL == "" {
			return apperr.Invalid("Missing required options `--alias` and `--url`.", "Example: db add --alias dev --url http://127.0.0.1:8090")
		}
		if err := d.dbStore.Add(*alias, *baseURL); err != nil {
			return mapStoreError(err)
		}
		_, _ = fmt.Fprintf(d.stdout, "Saved db alias %q.\n", *alias)
		return nil
	case "list":
		if len(args) > 1 {
			return apperr.Invalid("`db list` does not accept extra arguments.", "Use: db list")
		}
		items, err := d.dbStore.List()
		if err != nil {
			return mapStoreError(err)
		}
		rows := make([]map[string]any, 0, len(items))
		for _, it := range items {
			rows = append(rows, map[string]any{"db_alias": it.Alias, "base_url": it.BaseURL})
		}
		_, _ = fmt.Fprintln(d.stdout, renderTable([]string{"db_alias", "base_url"}, rows))
		return nil
	case "remove":
		fs := newFlagSet("db remove")
		alias := fs.String("alias", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *alias == "" {
			return apperr.Invalid("Missing required option `--alias`.", "Example: db remove --alias dev")
		}
		if err := d.dbStore.Remove(*alias); err != nil {
			return mapStoreError(err)
		}
		d.dropContextByDB(*alias)
		_, _ = fmt.Fprintf(d.stdout, "Removed db alias %q.\n", *alias)
		return nil
	default:
		return apperr.Invalid("Unknown db subcommand `"+cmd+"`.", "Use: db add|list|remove")
	}
}

func (d *Dispatcher) execSuperuser(args []string) error {
	if len(args) == 0 {
		return apperr.Invalid("Missing superuser subcommand.", "Use: superuser add|list|remove")
	}
	cmd := args[0]
	switch cmd {
	case "add":
		fs := newFlagSet("superuser add")
		dbAlias := fs.String("db", "", "")
		alias := fs.String("alias", "", "")
		email := fs.String("email", "", "")
		password := fs.String("password", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *dbAlias == "" || *alias == "" || *email == "" || *password == "" {
			return apperr.Invalid("Missing required options for superuser add.", "Example: superuser add --db dev --alias root --email admin@example.com --password secret")
		}
		if _, found, err := d.dbStore.Find(*dbAlias); err != nil {
			return mapStoreError(err)
		} else if !found {
			return apperr.Invalid("Could not find a saved db named \""+*dbAlias+"\".", "Run `pbviewer db list` to see available db aliases.")
		}
		if err := d.suStore.Add(*dbAlias, *alias, *email, *password); err != nil {
			return mapStoreError(err)
		}
		_, _ = fmt.Fprintf(d.stdout, "Saved superuser alias %q for db %q.\n", *alias, *dbAlias)
		return nil
	case "list":
		fs := newFlagSet("superuser list")
		dbAlias := fs.String("db", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *dbAlias == "" {
			return apperr.Invalid("Missing required option `--db`.", "Example: superuser list --db dev")
		}
		items, err := d.suStore.ListByDB(*dbAlias)
		if err != nil {
			return mapStoreError(err)
		}
		rows := make([]map[string]any, 0, len(items))
		for _, it := range items {
			rows = append(rows, map[string]any{"db_alias": it.DBAlias, "superuser_alias": it.Alias, "email": it.Email})
		}
		_, _ = fmt.Fprintln(d.stdout, renderTable([]string{"db_alias", "superuser_alias", "email"}, rows))
		return nil
	case "remove":
		fs := newFlagSet("superuser remove")
		dbAlias := fs.String("db", "", "")
		alias := fs.String("alias", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *dbAlias == "" || *alias == "" {
			return apperr.Invalid("Missing required options `--db` and `--alias`.", "Example: superuser remove --db dev --alias root")
		}
		if err := d.suStore.Remove(*dbAlias, *alias); err != nil {
			return mapStoreError(err)
		}
		d.dropContextBySuperuser(*dbAlias, *alias)
		_, _ = fmt.Fprintf(d.stdout, "Removed superuser alias %q from db %q.\n", *alias, *dbAlias)
		return nil
	default:
		return apperr.Invalid("Unknown superuser subcommand `"+cmd+"`.", "Use: superuser add|list|remove")
	}
}

func (d *Dispatcher) execContext(args []string) error {
	if len(args) == 0 {
		return apperr.Invalid("Missing context subcommand.", "Use: context show|use|save|clear|unsave")
	}

	sub := args[0]
	switch sub {
	case "show":
		if len(args) > 1 {
			return apperr.Invalid("`context show` does not accept extra arguments.", "Use: context show")
		}
		rows := []map[string]any{}
		if strings.TrimSpace(d.sessionCtx.DBAlias) != "" || strings.TrimSpace(d.sessionCtx.SuperuserAlias) != "" {
			rows = append(rows, map[string]any{
				"source":          "session",
				"db_alias":        d.sessionCtx.DBAlias,
				"superuser_alias": d.sessionCtx.SuperuserAlias,
			})
		}
		if d.hasSaved {
			rows = append(rows, map[string]any{
				"source":          "saved",
				"db_alias":        d.savedCtx.DBAlias,
				"superuser_alias": d.savedCtx.SuperuserAlias,
			})
		}
		if len(rows) == 0 {
			_, _ = fmt.Fprintln(d.stdout, "No context configured.")
			return nil
		}
		_, _ = fmt.Fprintln(d.stdout, renderTable([]string{"source", "db_alias", "superuser_alias"}, rows))
		return nil

	case "use":
		fs := newFlagSet("context use")
		dbAlias := fs.String("db", "", "")
		suAlias := fs.String("superuser", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if strings.TrimSpace(*dbAlias) == "" {
			return apperr.Invalid("Missing required option `--db`.", "Example: context use --db dev --superuser root")
		}
		db, found, err := d.dbStore.Find(*dbAlias)
		if err != nil {
			return mapStoreError(err)
		}
		if !found {
			return apperr.Invalid("Could not find a saved db named \""+*dbAlias+"\".", "Run `pbviewer db list` to see available db aliases.")
		}
		if strings.TrimSpace(*suAlias) != "" {
			if _, found, err := d.suStore.Find(db.Alias, *suAlias); err != nil {
				return mapStoreError(err)
			} else if !found {
				return apperr.Invalid("Superuser alias \""+*suAlias+"\" is not configured for db \""+db.Alias+"\".", "Run `pbviewer superuser list --db "+db.Alias+"` to see available aliases.")
			}
		}
		d.sessionCtx = commandContext{DBAlias: db.Alias, SuperuserAlias: strings.TrimSpace(*suAlias)}
		if d.sessionCtx.SuperuserAlias == "" {
			_, _ = fmt.Fprintf(d.stdout, "Updated session context: db=%q.\n", d.sessionCtx.DBAlias)
			return nil
		}
		_, _ = fmt.Fprintf(d.stdout, "Updated session context: db=%q superuser=%q.\n", d.sessionCtx.DBAlias, d.sessionCtx.SuperuserAlias)
		return nil

	case "save":
		if len(args) > 1 {
			return apperr.Invalid("`context save` does not accept extra arguments.", "Use: context save")
		}
		if strings.TrimSpace(d.sessionCtx.DBAlias) == "" {
			return apperr.Invalid("No session context to save.", "Run `context use --db <alias> [--superuser <alias>]` first.")
		}
		if err := d.persistSavedContext(d.sessionCtx); err != nil {
			return err
		}
		if d.sessionCtx.SuperuserAlias == "" {
			_, _ = fmt.Fprintf(d.stdout, "Saved default context: db=%q.\n", d.sessionCtx.DBAlias)
			return nil
		}
		_, _ = fmt.Fprintf(d.stdout, "Saved default context: db=%q superuser=%q.\n", d.sessionCtx.DBAlias, d.sessionCtx.SuperuserAlias)
		return nil

	case "clear":
		if len(args) > 1 {
			return apperr.Invalid("`context clear` does not accept extra arguments.", "Use: context clear")
		}
		d.sessionCtx = commandContext{}
		_, _ = fmt.Fprintln(d.stdout, "Cleared session context.")
		return nil

	case "unsave":
		if len(args) > 1 {
			return apperr.Invalid("`context unsave` does not accept extra arguments.", "Use: context unsave")
		}
		if err := d.ctxStore.Clear(); err != nil {
			return mapStoreError(err)
		}
		d.savedCtx = commandContext{}
		d.hasSaved = false
		_, _ = fmt.Fprintln(d.stdout, "Removed saved default context.")
		return nil

	default:
		return apperr.Invalid("Unknown context subcommand `"+sub+"`.", "Use: context show|use|save|clear|unsave")
	}
}

func (d *Dispatcher) execAPI(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return apperr.Invalid("Missing api subcommand.", "Use: api collections|collection|records|record")
	}

	sub := args[0]
	switch sub {
	case "collections":
		fs := newFlagSet("api collections")
		dbAlias := fs.String("db", "", "")
		suAlias := fs.String("superuser", "", "")
		format := fs.String("format", "table", "")
		out := fs.String("out", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if _, err := validateOutputOptions(*format, *out); err != nil {
			return err
		}
		target, err := d.resolveTarget(*dbAlias, *suAlias)
		if err != nil {
			return err
		}
		payload, err := d.getJSONWithAuth(ctx, target, pocketbase.BuildCollectionsEndpoint(), nil)
		if err != nil {
			return err
		}
		result := pocketbase.ParseItemsResult(payload)
		return d.writeQueryResult(*format, *out, result)

	case "collection":
		fs := newFlagSet("api collection")
		dbAlias := fs.String("db", "", "")
		suAlias := fs.String("superuser", "", "")
		name := fs.String("name", "", "")
		format := fs.String("format", "table", "")
		out := fs.String("out", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *name == "" {
			return apperr.Invalid("Missing required option `--name`.", "Example: api collection --name posts")
		}
		if _, err := validateOutputOptions(*format, *out); err != nil {
			return err
		}
		target, err := d.resolveTarget(*dbAlias, *suAlias)
		if err != nil {
			return err
		}
		payload, err := d.getJSONWithAuth(ctx, target, pocketbase.BuildCollectionEndpoint(*name), nil)
		if err != nil {
			return err
		}
		result := pocketbase.ParseSingleResult(payload)
		return d.writeQueryResult(*format, *out, result)

	case "records":
		fs := newFlagSet("api records")
		dbAlias := fs.String("db", "", "")
		suAlias := fs.String("superuser", "", "")
		collection := fs.String("collection", "", "")
		page := fs.String("page", "", "")
		perPage := fs.String("per-page", "", "")
		sortExpr := fs.String("sort", "", "")
		filterExpr := fs.String("filter", "", "")
		format := fs.String("format", "table", "")
		out := fs.String("out", "", "")
		view := fs.String("view", "auto", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *collection == "" {
			return apperr.Invalid("Missing required option `--collection`.", "Example: api records --collection posts")
		}
		normalizedFormat, err := validateOutputOptions(*format, *out)
		if err != nil {
			return err
		}
		normalizedView, err := normalizeView(*view)
		if err != nil {
			return err
		}
		if normalizedFormat != "table" && normalizedView == "tui" {
			return apperr.Invalid("`--view tui` requires `--format table`.", "Use `--format table` or switch view to `table`/`auto`.")
		}
		state := RecordsQueryState{
			Collection: *collection,
			Sort:       *sortExpr,
			Filter:     *filterExpr,
		}
		if *page != "" {
			v, err := positiveInt(*page)
			if err != nil {
				return apperr.Invalid("Invalid `--page` value.", "`--page` must be a positive integer.")
			}
			state.Page = v
		}
		if *perPage != "" {
			v, err := positiveInt(*perPage)
			if err != nil {
				return apperr.Invalid("Invalid `--per-page` value.", "`--per-page` must be a positive integer.")
			}
			state.PerPage = v
		}
		target, err := d.resolveTarget(*dbAlias, *suAlias)
		if err != nil {
			return err
		}

		shouldTUI := false
		switch normalizedView {
		case "tui":
			if normalizedFormat != "table" {
				return apperr.Invalid("`--view tui` requires `--format table`.", "Use `--format table`.")
			}
			if !d.IsInteractiveTTY() {
				return apperr.Invalid("`--view tui` requires interactive REPL TTY mode.", "Run `pbviewer` in a terminal and execute this command there.")
			}
			shouldTUI = true
		case "auto":
			shouldTUI = normalizedFormat == "table" && d.IsInteractiveTTY()
		case "table":
			shouldTUI = false
		}

		if shouldTUI {
			return d.runRecordsTUI(ctx, target, state)
		}
		result, err := d.fetchRecords(ctx, target, state)
		if err != nil {
			return err
		}
		return d.writeQueryResult(normalizedFormat, *out, result)

	case "record":
		fs := newFlagSet("api record")
		dbAlias := fs.String("db", "", "")
		suAlias := fs.String("superuser", "", "")
		collection := fs.String("collection", "", "")
		recordID := fs.String("id", "", "")
		format := fs.String("format", "table", "")
		out := fs.String("out", "", "")
		if err := fs.Parse(args[1:]); err != nil {
			return invalidFlagError(err)
		}
		if *collection == "" || *recordID == "" {
			return apperr.Invalid("Missing required options `--collection` and `--id`.", "Example: api record --collection posts --id rec123")
		}
		normalizedFormat, err := validateOutputOptions(*format, *out)
		if err != nil {
			return err
		}
		target, err := d.resolveTarget(*dbAlias, *suAlias)
		if err != nil {
			return err
		}
		payload, err := d.getJSONWithAuth(ctx, target, pocketbase.BuildRecordEndpoint(*collection, *recordID), nil)
		if err != nil {
			return err
		}
		result := pocketbase.ParseSingleResult(payload)
		return d.writeQueryResult(normalizedFormat, *out, result)
	default:
		return apperr.Invalid("This CLI is read-only for PocketBase API operations.", "Only GET requests are supported.")
	}
}

func normalizeView(view string) (string, error) {
	v := strings.ToLower(strings.TrimSpace(view))
	if v == "" {
		return "auto", nil
	}
	switch v {
	case "auto", "tui", "table":
		return v, nil
	default:
		return "", apperr.Invalid("Unsupported view mode.", "Use one of: auto, tui, table.")
	}
}

type pbTarget struct {
	DB storage.DB
	SU storage.Superuser
}

func (d *Dispatcher) resolveTarget(dbAlias, suAlias string) (pbTarget, error) {
	resolvedDB, resolvedSU, err := d.resolveAliases(dbAlias, suAlias)
	if err != nil {
		return pbTarget{}, err
	}

	db, found, err := d.dbStore.Find(resolvedDB)
	if err != nil {
		return pbTarget{}, mapStoreError(err)
	}
	if !found {
		return pbTarget{}, apperr.Invalid("Could not find a saved db named \""+resolvedDB+"\".", "Run `pbviewer db list` to see available db aliases.")
	}
	su, found, err := d.suStore.Find(db.Alias, resolvedSU)
	if err != nil {
		return pbTarget{}, mapStoreError(err)
	}
	if !found {
		return pbTarget{}, apperr.Invalid("Superuser alias \""+resolvedSU+"\" is not configured for db \""+db.Alias+"\".", "Run `pbviewer superuser list --db "+db.Alias+"` to see available aliases.")
	}
	return pbTarget{DB: db, SU: su}, nil
}

func (d *Dispatcher) resolveAliases(dbAlias, suAlias string) (string, string, error) {
	explicit := commandContext{DBAlias: strings.TrimSpace(dbAlias), SuperuserAlias: strings.TrimSpace(suAlias)}

	resolvedDB := explicit.DBAlias
	if resolvedDB == "" {
		if strings.TrimSpace(d.sessionCtx.DBAlias) != "" {
			resolvedDB = d.sessionCtx.DBAlias
		} else if d.hasSaved {
			resolvedDB = d.savedCtx.DBAlias
		}
	}

	resolvedSU := explicit.SuperuserAlias
	if resolvedSU == "" {
		if strings.TrimSpace(d.sessionCtx.SuperuserAlias) != "" && contextMatchesDB(d.sessionCtx, resolvedDB) {
			resolvedSU = d.sessionCtx.SuperuserAlias
		} else if d.hasSaved && strings.TrimSpace(d.savedCtx.SuperuserAlias) != "" && contextMatchesDB(d.savedCtx, resolvedDB) {
			resolvedSU = d.savedCtx.SuperuserAlias
		}
	}

	if strings.TrimSpace(resolvedDB) == "" || strings.TrimSpace(resolvedSU) == "" {
		return "", "", apperr.Invalid("Missing required options `--db` and `--superuser`.", "Set context with `context use --db <alias> --superuser <alias>` or provide flags explicitly.")
	}
	return resolvedDB, resolvedSU, nil
}

func contextMatchesDB(ctx commandContext, resolvedDB string) bool {
	if strings.TrimSpace(ctx.DBAlias) == "" {
		return true
	}
	if strings.TrimSpace(resolvedDB) == "" {
		return false
	}
	return strings.EqualFold(ctx.DBAlias, resolvedDB)
}

func (d *Dispatcher) fetchRecords(ctx context.Context, target pbTarget, state RecordsQueryState) (pocketbase.QueryResult, error) {
	payload, err := d.getJSONWithAuth(ctx, target, pocketbase.BuildRecordsEndpoint(state.Collection), state.QueryParams())
	if err != nil {
		return pocketbase.QueryResult{}, err
	}
	return pocketbase.ParseItemsResult(payload), nil
}

func (d *Dispatcher) getJSONWithAuth(ctx context.Context, target pbTarget, endpoint string, query map[string]string) (map[string]any, error) {
	token, err := d.authenticate(ctx, target, false)
	if err != nil {
		return nil, err
	}

	payload, err := d.pbClient.GetJSON(ctx, target.DB.BaseURL, token, endpoint, query)
	if err == nil {
		return payload, nil
	}
	var authErr *pocketbase.AuthError
	if !errors.As(err, &authErr) {
		return nil, mapPBError(err, target.SU.Alias, target.DB.Alias)
	}

	d.clearAuthCache(target)
	token, err = d.authenticate(ctx, target, true)
	if err != nil {
		return nil, err
	}
	payload, err = d.pbClient.GetJSON(ctx, target.DB.BaseURL, token, endpoint, query)
	if err != nil {
		return nil, mapPBError(err, target.SU.Alias, target.DB.Alias)
	}
	return payload, nil
}

func (d *Dispatcher) authenticate(ctx context.Context, target pbTarget, force bool) (string, error) {
	if !force {
		if token, ok := d.getCachedToken(target); ok {
			return token, nil
		}
	}
	token, err := d.pbClient.Authenticate(ctx, target.DB.BaseURL, target.SU.Email, target.SU.Password)
	if err != nil {
		return "", mapPBError(err, target.SU.Alias, target.DB.Alias)
	}
	d.storeCachedToken(target, token)
	return token, nil
}

func (d *Dispatcher) getCachedToken(target pbTarget) (string, bool) {
	entry, ok := d.authCache[authCacheKey{dbAlias: strings.ToLower(target.DB.Alias), suAlias: strings.ToLower(target.SU.Alias)}]
	if !ok {
		return "", false
	}
	if entry.hasExpiry {
		now := d.now().UTC()
		if !entry.expiresAt.After(now.Add(30 * time.Second)) {
			d.clearAuthCache(target)
			return "", false
		}
	}
	return entry.token, true
}

func (d *Dispatcher) storeCachedToken(target pbTarget, token string) {
	entry := authCacheEntry{token: token}
	if expiresAt, ok := parseTokenExpiry(token); ok {
		entry.expiresAt = expiresAt
		entry.hasExpiry = true
	}
	key := authCacheKey{dbAlias: strings.ToLower(target.DB.Alias), suAlias: strings.ToLower(target.SU.Alias)}
	d.authCache[key] = entry
}

func (d *Dispatcher) clearAuthCache(target pbTarget) {
	delete(d.authCache, authCacheKey{dbAlias: strings.ToLower(target.DB.Alias), suAlias: strings.ToLower(target.SU.Alias)})
}

func (d *Dispatcher) writeQueryResult(format, out string, result pocketbase.QueryResult) error {
	text, err := RenderQueryResult(format, out, result)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(d.stdout, text)
	return nil
}

func (d *Dispatcher) persistSavedContext(ctx commandContext) error {
	err := d.ctxStore.Save(storage.Context{DBAlias: ctx.DBAlias, SuperuserAlias: ctx.SuperuserAlias})
	if err != nil {
		return mapStoreError(err)
	}
	d.savedCtx = ctx
	d.hasSaved = true
	return nil
}

func (d *Dispatcher) dropContextByDB(dbAlias string) {
	if strings.EqualFold(d.sessionCtx.DBAlias, dbAlias) {
		d.sessionCtx = commandContext{}
	}
	if d.hasSaved && strings.EqualFold(d.savedCtx.DBAlias, dbAlias) {
		_ = d.ctxStore.Clear()
		d.savedCtx = commandContext{}
		d.hasSaved = false
	}
}

func (d *Dispatcher) dropContextBySuperuser(dbAlias, suAlias string) {
	if strings.EqualFold(d.sessionCtx.DBAlias, dbAlias) && strings.EqualFold(d.sessionCtx.SuperuserAlias, suAlias) {
		d.sessionCtx.SuperuserAlias = ""
	}
	if d.hasSaved && strings.EqualFold(d.savedCtx.DBAlias, dbAlias) && strings.EqualFold(d.savedCtx.SuperuserAlias, suAlias) {
		d.savedCtx.SuperuserAlias = ""
		_ = d.persistSavedContext(d.savedCtx)
	}
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func invalidFlagError(err error) error {
	if err == nil {
		return nil
	}
	return apperr.Invalid("Invalid command arguments.", err.Error())
}

func mapStoreError(err error) error {
	if err == nil {
		return nil
	}
	var validationErr *storage.ValidationError
	if errors.As(err, &validationErr) {
		return apperr.Invalid(validationErr.Message, "")
	}
	return apperr.RuntimeErr("Local configuration storage failed.", "Check local file permissions and retry.", err)
}

func mapPBError(err error, superuserAlias, dbAlias string) error {
	if err == nil {
		return nil
	}
	var authErr *pocketbase.AuthError
	if errors.As(err, &authErr) {
		return apperr.ExternalErr("Authentication failed for superuser \""+superuserAlias+"\" on db \""+dbAlias+"\".", "Verify the saved credentials for this superuser alias.", err)
	}
	if pocketbase.IsNetworkError(err) {
		return apperr.ExternalErr("Network request to PocketBase failed.", "Check db URL and network connectivity.", err)
	}
	var apiErr *pocketbase.APIError
	if errors.As(err, &apiErr) {
		return apperr.ExternalErr(fmt.Sprintf("PocketBase API request failed with status %d.", apiErr.Status), "Check credentials, query parameters, and target resource.", err)
	}
	return apperr.ExternalErr("PocketBase request failed.", "Check connectivity and server status.", err)
}

func positiveInt(s string) (int, error) {
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	if v <= 0 {
		return 0, fmt.Errorf("must be positive")
	}
	return v, nil
}

func (d *Dispatcher) printHelp() {
	help := strings.TrimSpace(`pbviewer command reference

Run modes:
  pbviewer                         Start REPL mode.
  pbviewer -c "<command>"          Run one command and exit.
  pbviewer <script-file>           Execute commands from a script file.

Core commands:
  version                         Print CLI version.
  help                            Show available commands.

DB commands:
  db add --alias <dbAlias> --url <baseUrl>
                                  Save a PocketBase base URL as a db alias.
  db list                         List saved db aliases.
  db remove --alias <dbAlias>     Remove a saved db alias.

Superuser commands:
  superuser add --db <dbAlias> --alias <superuserAlias> --email <email> --password <password>
                                  Save superuser credentials for a db alias.
  superuser list --db <dbAlias>   List superuser aliases for a db alias.
  superuser remove --db <dbAlias> --alias <superuserAlias>
                                  Remove a saved superuser alias.

Context commands:
  context show                    Show current session/saved target context.
  context use --db <dbAlias> [--superuser <superuserAlias>]
                                  Set active session target context.
  context save                    Save current session context as default.
  context clear                   Clear current session context.
  context unsave                  Remove saved default context.

API commands (read-only GET):
  api collections --db <dbAlias> --superuser <superuserAlias> [--format table|csv|markdown] [--out <path>]
                                  List collections from PocketBase.
  api collection --db <dbAlias> --superuser <superuserAlias> --name <collectionName> [--format table|csv|markdown] [--out <path>]
                                  Get one collection by name.
  api records --db <dbAlias> --superuser <superuserAlias> --collection <collectionName> [--page <n>] [--per-page <n>] [--sort <expr>] [--filter <expr>] [--view auto|tui|table] [--format table|csv|markdown] [--out <path>]
                                  List records with paging, sort, and filter options.
  api record --db <dbAlias> --superuser <superuserAlias> --collection <collectionName> --id <recordId> [--format table|csv|markdown] [--out <path>]
                                  Get one record by id.

Output:
  Default format is table.
  csv/markdown requires --out <path>.
  TUI view is available in interactive REPL TTY mode.`)
	_, _ = fmt.Fprintln(d.stdout, help)
}
