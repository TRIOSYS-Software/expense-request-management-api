package repositories_test

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"

	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/storage"
	"shwetaik-expense-management-api/tests/testutil"

	"gorm.io/gorm"
)

var updateGolden = flag.Bool("update", false, "rewrite the golden files in testdata/")

func assertGolden(t *testing.T, name string, got any) {
	t.Helper()

	encoded, err := json.MarshalIndent(got, "", "  ")
	if err != nil {
		t.Fatalf("marshal %s: %v", name, err)
	}
	encoded = append(encoded, '\n')

	path := filepath.Join("testdata", name+".json")

	if *updateGolden {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("create testdata dir: %v", err)
		}
		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			t.Fatalf("write golden %s: %v", path, err)
		}
		t.Logf("golden updated: %s", path)
		return
	}

	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read golden %s (run with -update to create): %v", path, err)
	}
	if string(want) != string(encoded) {
		t.Errorf("summary for %s drifted from its golden file.\n--- want ---\n%s\n--- got ---\n%s",
			name, want, encoded)
	}
}

func day(t *testing.T, s string) time.Time {
	t.Helper()
	parsed, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		t.Fatalf("parse date %q: %v", s, err)
	}
	return parsed.Add(10 * time.Hour)
}
func seedSummaryWorld(t *testing.T, db *gorm.DB) (testutil.Baseline, uint) {
	t.Helper()
	b := testutil.SeedBaseline(t, db)

	other := b.NewUser(t, db, "Second Submitter", "second@example.com")
	ar := b.NewAdvance(t, db, 1000, "approved")

	// Baseline user, 2024-03-01.
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 100, Status: "pending", DateSubmitted: testutil.Ptr(day(t, "2024-03-01")),
	})
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 200, Status: "approved", DateSubmitted: testutil.Ptr(day(t, "2024-03-01")),
	})

	// Baseline user, 2024-03-02 — one of them draws on the advance.
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 300, Status: "approved", DateSubmitted: testutil.Ptr(day(t, "2024-03-02")),
		AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(400.0),
	})
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 50, Status: "rejected", DateSubmitted: testutil.Ptr(day(t, "2024-03-02")),
		AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(50.0),
	})

	// Second user, 2024-03-05, plus a completed one.
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 500, Status: "completed", DateSubmitted: testutil.Ptr(day(t, "2024-03-05")),
		UserID: &other.ID,
	})
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 75, Status: "pending", DateSubmitted: testutil.Ptr(day(t, "2024-03-05")),
		UserID: &other.ID,
	})

	// Outside the 03-01..03-05 window used by the dated cases.
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 900, Status: "approved", DateSubmitted: testutil.Ptr(day(t, "2024-04-20")),
	})

	return b, other.ID
}

func TestSummaryExpenseGolden(t *testing.T) {
	db := testutil.DB(t)

	cases := []struct {
		name    string
		filters func(b testutil.Baseline, otherID uint) map[string]any
	}{
		{
			name:    "expense_no_filters",
			filters: func(testutil.Baseline, uint) map[string]any { return map[string]any{} },
		},
		{
			name: "expense_by_user",
			filters: func(b testutil.Baseline, _ uint) map[string]any {
				return map[string]any{"user_id": b.User.ID}
			},
		},
		{
			name: "expense_status_approved",
			filters: func(testutil.Baseline, uint) map[string]any {
				return map[string]any{"status": "approved"}
			},
		},
		{
			name: "expense_status_pending",
			filters: func(testutil.Baseline, uint) map[string]any {
				return map[string]any{"status": "pending"}
			},
		},
		{
			name: "expense_date_range",
			filters: func(testutil.Baseline, uint) map[string]any {
				return map[string]any{"start_date": "2024-03-01", "end_date": "2024-03-05"}
			},
		},
		{
			name: "expense_date_range_and_user",
			filters: func(b testutil.Baseline, _ uint) map[string]any {
				return map[string]any{
					"start_date": "2024-03-01", "end_date": "2024-03-05",
					"user_id": b.User.ID,
				}
			},
		},
		{
			name: "expense_search_by_submitter_name",
			filters: func(testutil.Baseline, uint) map[string]any {
				return map[string]any{"search": "Second"}
			},
		},
		{
			name: "expense_search_by_project_code",
			filters: func(testutil.Baseline, uint) map[string]any {
				return map[string]any{"search": "PRJ-1"}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b, otherID := seedSummaryWorld(t, db)
			repo := repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir()))

			got, err := repo.GetExpenseRequestsSummary(tc.filters(b, otherID))
			if err != nil {
				t.Fatalf("GetExpenseRequestsSummary: %v", err)
			}
			assertGolden(t, tc.name, got)
		})
	}
}

func TestSummaryAdvanceGolden(t *testing.T) {
	db := testutil.DB(t)

	seed := func(t *testing.T, db *gorm.DB) testutil.Baseline {
		t.Helper()
		b := testutil.SeedBaseline(t, db)

		open := b.NewAdvance(t, db, 1000, "approved")
		b.NewExpense(t, db, testutil.ExpenseOpts{
			Amount: 400, Status: "approved", DateSubmitted: testutil.Ptr(day(t, "2024-03-02")),
			AdvanceRequestID: &open.ID, AdvanceUsed: testutil.Ptr(400.0),
		})

		untouched := b.NewAdvance(t, db, 750, "approved")
		_ = untouched

		b.NewAdvance(t, db, 250, "pending")
		b.NewAdvance(t, db, 300, "rejected")
		b.NewAdvance(t, db, 500, "completed")

		return b
	}

	cases := []struct {
		name    string
		filters func(b testutil.Baseline) map[string]any
	}{
		{
			name:    "advance_no_filters",
			filters: func(testutil.Baseline) map[string]any { return map[string]any{} },
		},
		{
			name: "advance_by_user",
			filters: func(b testutil.Baseline) map[string]any {
				return map[string]any{"user_id": b.User.ID}
			},
		},
		{
			name: "advance_status_approved",
			filters: func(testutil.Baseline) map[string]any {
				return map[string]any{"status": "approved"}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b := seed(t, db)
			repo := repositories.NewAdvanceRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir()))

			got, err := repo.GetAdvanceRequestsSummary(tc.filters(b))
			if err != nil {
				t.Fatalf("GetAdvanceRequestsSummary: %v", err)
			}
			assertGolden(t, tc.name, got)
		})
	}
}
