package cli

import (
	"bytes"
	"context"
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/jiseop121/pbdash/internal/pocketbase"
	"github.com/jiseop121/pbdash/internal/storage"
)

func TestNavigatorTUIHandleKeySupportsArrowHorizontalNavigation(t *testing.T) {
	ui := &navigatorTUI{
		screen:    screenRecords,
		tableView: tview.NewTable(),
		result:    queryResultWithColumns(visibleColumnWindow + 2),
	}

	ui.handleKey(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
	if ui.columnOffset != 1 {
		t.Fatalf("column offset mismatch after right: got=%d want=1", ui.columnOffset)
	}

	ui.handleKey(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
	if ui.columnOffset != 0 {
		t.Fatalf("column offset mismatch after left: got=%d want=0", ui.columnOffset)
	}
}

func TestNavigatorTUIArrowHorizontalNavigationBounds(t *testing.T) {
	ui := &navigatorTUI{
		screen:    screenRecords,
		tableView: tview.NewTable(),
		result:    queryResultWithColumns(visibleColumnWindow + 1),
	}

	ui.handleKey(tcell.NewEventKey(tcell.KeyLeft, 0, tcell.ModNone))
	if ui.columnOffset != 0 {
		t.Fatalf("left at start should stay 0, got=%d", ui.columnOffset)
	}

	for i := 0; i < visibleColumnWindow+4; i++ {
		ui.handleKey(tcell.NewEventKey(tcell.KeyRight, 0, tcell.ModNone))
	}
	if ui.columnOffset != 1 {
		t.Fatalf("right bound mismatch: got=%d want=1", ui.columnOffset)
	}
}

func TestPickPreferredSuperuser(t *testing.T) {
	candidates := []storage.Superuser{
		{DBAlias: "dev", Alias: "ops", Email: "ops@example.com"},
		{DBAlias: "dev", Alias: "root", Email: "root@example.com"},
	}

	got, ok := pickPreferredSuperuser("dev", candidates, commandContext{DBAlias: "dev", SuperuserAlias: "root"}, commandContext{})
	if !ok || got.Alias != "root" {
		t.Fatalf("expected session context superuser, got=%+v ok=%v", got, ok)
	}

	got, ok = pickPreferredSuperuser("dev", candidates, commandContext{}, commandContext{DBAlias: "dev", SuperuserAlias: "ops"})
	if !ok || got.Alias != "ops" {
		t.Fatalf("expected saved context superuser, got=%+v ok=%v", got, ok)
	}

	got, ok = pickPreferredSuperuser("dev", candidates[:1], commandContext{}, commandContext{})
	if !ok || got.Alias != "ops" {
		t.Fatalf("expected single configured superuser, got=%+v ok=%v", got, ok)
	}
}

func TestRunRecordsTUIUsesNavigatorRunner(t *testing.T) {
	d := NewDispatcher(DispatcherConfig{Stdout: bytes.NewBuffer(nil), Version: "test", DataDir: t.TempDir()})
	target := pbTarget{
		DB: storage.DB{Alias: "dev", BaseURL: "http://127.0.0.1:8090"},
		SU: storage.Superuser{DBAlias: "dev", Alias: "root", Email: "root@example.com"},
	}
	state := RecordsQueryState{Collection: "posts", Page: 2}

	var gotRoute navigatorRoute
	d.navigatorRunner = func(_ context.Context, route navigatorRoute) error {
		gotRoute = route
		return nil
	}

	if err := d.runRecordsTUI(context.Background(), target, state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !gotRoute.hasTarget || gotRoute.screen != screenRecords {
		t.Fatalf("unexpected route: %+v", gotRoute)
	}
	if gotRoute.target.DB.Alias != "dev" || gotRoute.target.SU.Alias != "root" {
		t.Fatalf("target mismatch: %+v", gotRoute.target)
	}
	if gotRoute.state.Collection != "posts" || gotRoute.state.Page != 2 {
		t.Fatalf("state mismatch: %+v", gotRoute.state)
	}
}

func TestNavigatorTUISetupViewsCapturesTableShortcuts(t *testing.T) {
	ui := &navigatorTUI{
		app:           tview.NewApplication(),
		statusView:    tview.NewTextView(),
		tableView:     tview.NewTable(),
		detailView:    tview.NewTextView(),
		helpView:      tview.NewTextView(),
		screen:        screenRecords,
		detailVisible: true,
		observedCols:  map[string]struct{}{},
		result: pocketbase.QueryResult{Rows: []map[string]any{
			{"id": "1", "title": "first"},
			{"id": "2", "title": "second"},
		}},
	}

	ui.setupViews()
	handler := ui.tableView.InputHandler()
	require.NotNil(t, handler)

	handler(tcell.NewEventKey(tcell.KeyRune, 'j', tcell.ModNone), func(tview.Primitive) {})
	assert.Equal(t, 1, ui.selectedIndex)

	handler(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone), func(tview.Primitive) {})
	assert.False(t, ui.detailVisible)
}

func queryResultWithColumns(count int) pocketbaseQueryResult {
	rows := []map[string]any{rowWithColumns(count)}
	return pocketbase.QueryResult{Rows: rows}
}

type pocketbaseQueryResult = pocketbase.QueryResult

func rowWithColumns(count int) map[string]any {
	row := make(map[string]any, count)
	for i := 0; i < count; i++ {
		row[fmt.Sprintf("c%02d", i)] = i
	}
	return row
}
