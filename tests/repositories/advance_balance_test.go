package repositories_test

// Characterization tests for the advance-settlement arithmetic.

import (
	"testing"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/tests/testutil"

	"gorm.io/gorm"
)

const eps = 1e-9

func assertFloat(t *testing.T, label string, got, want float64) {
	t.Helper()
	if diff := got - want; diff > eps || diff < -eps {
		t.Errorf("%s = %v, want %v", label, got, want)
	}
}

func TestAdvanceConsumed(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)

	tests := []struct {
		name     string
		expenses []testutil.ExpenseOpts
		statuses []string
		want     float64
	}{
		{
			name:     "no linked expenses consumes nothing",
			expenses: nil,
			statuses: []string{"pending", "approved"},
			want:     0,
		},
		{
			name: "expense without advance_used_amount charges its full amount",
			expenses: []testutil.ExpenseOpts{
				{Amount: 400, Status: "approved"},
			},
			statuses: []string{"pending", "approved"},
			want:     400,
		},
		{
			name: "spending less than drawn charges only what was spent",
			expenses: []testutil.ExpenseOpts{
				{Amount: 300, Status: "approved", AdvanceUsed: testutil.Ptr(500.0)},
			},
			statuses: []string{"pending", "approved"},
			want:     300,
		},
		{
			name: "returning the unspent remainder charges the whole draw",
			expenses: []testutil.ExpenseOpts{
				{Amount: 300, Status: "approved", AdvanceUsed: testutil.Ptr(500.0), Returned: testutil.Ptr(200.0)},
			},
			statuses: []string{"pending", "approved"},
			want:     500,
		},
		{
			name: "spending beyond the draw is capped at the draw",
			expenses: []testutil.ExpenseOpts{
				{Amount: 900, Status: "approved", AdvanceUsed: testutil.Ptr(500.0)},
			},
			statuses: []string{"pending", "approved"},
			want:     500,
		},
		{
			name: "pending and approved both count",
			expenses: []testutil.ExpenseOpts{
				{Amount: 200, Status: "approved", AdvanceUsed: testutil.Ptr(200.0)},
				{Amount: 300, Status: "pending", AdvanceUsed: testutil.Ptr(300.0)},
			},
			statuses: []string{"pending", "approved"},
			want:     500,
		},
		{
			name: "rejected expenses never count",
			expenses: []testutil.ExpenseOpts{
				{Amount: 200, Status: "approved", AdvanceUsed: testutil.Ptr(200.0)},
				{Amount: 300, Status: "rejected", AdvanceUsed: testutil.Ptr(300.0)},
			},
			statuses: []string{"pending", "approved"},
			want:     200,
		},
		{
			name: "approved-only scope excludes pending",
			expenses: []testutil.ExpenseOpts{
				{Amount: 200, Status: "approved", AdvanceUsed: testutil.Ptr(200.0)},
				{Amount: 300, Status: "pending", AdvanceUsed: testutil.Ptr(300.0)},
			},
			statuses: []string{"approved"},
			want:     200,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b = testutil.SeedBaseline(t, db)
			ar := b.NewAdvance(t, db, 1000, "approved")

			for _, e := range tc.expenses {
				e.AdvanceRequestID = &ar.ID
				b.NewExpense(t, db, e)
			}

			got, err := repositories.AdvanceConsumed(db, ar.ID, nil, tc.statuses)
			if err != nil {
				t.Fatalf("repositories.AdvanceConsumed: %v", err)
			}
			assertFloat(t, "consumed", got, tc.want)
		})
	}
}

func TestAdvanceConsumedExcludesGivenExpense(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	ar := b.NewAdvance(t, db, 1000, "approved")

	keep := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 200, Status: "approved",
		AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(200.0),
	})
	skip := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 300, Status: "approved",
		AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(300.0),
	})
	_ = keep

	got, err := repositories.AdvanceConsumed(db, ar.ID, &skip.ID, []string{"pending", "approved"})
	if err != nil {
		t.Fatalf("repositories.AdvanceConsumed: %v", err)
	}
	assertFloat(t, "consumed excluding one expense", got, 200)
}

func TestAdvanceRemainingSnapsSubKyatDustToZero(t *testing.T) {
	db := testutil.DB(t)

	tests := []struct {
		name        string
		advance     float64
		expenseAmt  float64
		wantRemains float64
	}{
		{name: "ordinary remainder is preserved", advance: 1000, expenseAmt: 400, wantRemains: 600},
		{name: "exactly one Kyat left is a real balance", advance: 1000, expenseAmt: 999, wantRemains: 1},
		{name: "sub-Kyat dust snaps to zero", advance: 1000, expenseAmt: 999.5, wantRemains: 0},
		{name: "exact settlement is zero", advance: 1000, expenseAmt: 1000, wantRemains: 0},
		{name: "over-consumption clamps to zero, never negative", advance: 1000, expenseAmt: 1500, wantRemains: 0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b := testutil.SeedBaseline(t, db)
			ar := b.NewAdvance(t, db, tc.advance, "approved")
			b.NewExpense(t, db, testutil.ExpenseOpts{
				Amount: tc.expenseAmt, Status: "approved", AdvanceRequestID: &ar.ID,
			})

			got, err := repositories.AdvanceRemaining(db, &ar, nil)
			if err != nil {
				t.Fatalf("repositories.AdvanceRemaining: %v", err)
			}
			assertFloat(t, "remaining", got, tc.wantRemains)
		})
	}
}

func TestFillAdvanceBalances(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)

	partial := b.NewAdvance(t, db, 1000, "approved")
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 400, Status: "approved", AdvanceRequestID: &partial.ID,
	})

	dusty := b.NewAdvance(t, db, 1000, "approved")
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 999.5, Status: "approved", AdvanceRequestID: &dusty.ID,
	})

	untouched := b.NewAdvance(t, db, 750, "approved")

	ars := []models.AdvanceRequests{partial, dusty, untouched}
	if err := repositories.FillAdvanceBalances(db, ars); err != nil {
		t.Fatalf("repositories.FillAdvanceBalances: %v", err)
	}

	assertFloat(t, "partial settled", ars[0].SettledAmount, 400)
	assertFloat(t, "partial remaining", ars[0].RemainingBalance, 600)

	// Dust is absorbed into settled rather than left as a 0.5 remainder.
	assertFloat(t, "dusty settled", ars[1].SettledAmount, 1000)
	assertFloat(t, "dusty remaining", ars[1].RemainingBalance, 0)

	assertFloat(t, "untouched settled", ars[2].SettledAmount, 0)
	assertFloat(t, "untouched remaining", ars[2].RemainingBalance, 750)
}

func TestFillAdvanceBalancesEmptySliceIsNoop(t *testing.T) {
	db := testutil.DB(t)
	if err := repositories.FillAdvanceBalances(db, nil); err != nil {
		t.Fatalf("repositories.FillAdvanceBalances(nil): %v", err)
	}
}

func TestAdvanceFullySettledCountsApprovedOnly(t *testing.T) {
	db := testutil.DB(t)

	tests := []struct {
		name   string
		status string
		amount float64
		want   bool
	}{
		{name: "approved expense covering the advance settles it", status: "approved", amount: 1000, want: true},
		{name: "approved expense leaving dust settles it", status: "approved", amount: 999.5, want: true},
		{name: "approved expense leaving a real balance does not", status: "approved", amount: 999, want: false},
		{name: "pending expense does not settle", status: "pending", amount: 1000, want: false},
		{name: "rejected expense does not settle", status: "rejected", amount: 1000, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b := testutil.SeedBaseline(t, db)
			ar := b.NewAdvance(t, db, 1000, "approved")
			b.NewExpense(t, db, testutil.ExpenseOpts{
				Amount: tc.amount, Status: tc.status, AdvanceRequestID: &ar.ID,
			})

			got, err := repositories.AdvanceFullySettled(db, &ar)
			if err != nil {
				t.Fatalf("repositories.AdvanceFullySettled: %v", err)
			}
			if got != tc.want {
				t.Errorf("repositories.AdvanceFullySettled = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestReconcileAdvanceStatuses(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)

	settled := b.NewAdvance(t, db, 1000, "approved")
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 1000, Status: "approved", AdvanceRequestID: &settled.ID,
	})

	outstanding := b.NewAdvance(t, db, 1000, "approved")
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 200, Status: "approved", AdvanceRequestID: &outstanding.ID,
	})

	// A pending advance is out of scope even when fully covered.
	pending := b.NewAdvance(t, db, 500, "pending")
	b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 500, Status: "approved", AdvanceRequestID: &pending.ID,
	})

	changed, err := repositories.ReconcileAdvanceStatuses(db)
	if err != nil {
		t.Fatalf("repositories.ReconcileAdvanceStatuses: %v", err)
	}
	if changed != 1 {
		t.Errorf("changed = %d, want 1", changed)
	}

	assertStatus(t, db, settled.ID, "completed")
	assertStatus(t, db, outstanding.ID, "approved")
	assertStatus(t, db, pending.ID, "pending")
}

func TestReconcileAdvanceStatusesCountsSoftDeletedExpenses(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)

	ar := b.NewAdvance(t, db, 1000, "approved")
	er := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 1000, Status: "approved", AdvanceRequestID: &ar.ID,
	})

	if err := db.Delete(&models.ExpenseRequests{}, er.ID).Error; err != nil {
		t.Fatalf("soft-delete expense: %v", err)
	}

	remaining, err := repositories.AdvanceRemaining(db, &ar, nil)
	if err != nil {
		t.Fatalf("repositories.AdvanceRemaining: %v", err)
	}
	assertFloat(t, "remaining ignores soft-deleted expense", remaining, 1000)

	changed, err := repositories.ReconcileAdvanceStatuses(db)
	if err != nil {
		t.Fatalf("repositories.ReconcileAdvanceStatuses: %v", err)
	}
	if changed != 1 {
		t.Errorf("changed = %d, want 1 (reconcile counts soft-deleted rows)", changed)
	}
	assertStatus(t, db, ar.ID, "completed")
}

func assertStatus(t *testing.T, db *gorm.DB, id uint, want string) {
	t.Helper()
	var got models.AdvanceRequests
	if err := db.First(&got, id).Error; err != nil {
		t.Fatalf("load advance %d: %v", id, err)
	}
	if got.Status != want {
		t.Errorf("advance %d status = %q, want %q", id, got.Status, want)
	}
}
