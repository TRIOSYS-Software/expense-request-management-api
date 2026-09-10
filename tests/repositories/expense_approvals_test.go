package repositories_test

import (
	"strings"
	"testing"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/tests/testutil"

	"gorm.io/gorm"
)

func newExpenseApprovalsRepo(db *gorm.DB) *repositories.ExpenseApprovalsRepo {
	return repositories.NewExpenseApprovalsRepo(db, nil)
}

func loadExpense(t *testing.T, db *gorm.DB, id uint) models.ExpenseRequests {
	t.Helper()
	var er models.ExpenseRequests
	if err := db.First(&er, id).Error; err != nil {
		t.Fatalf("load expense %d: %v", id, err)
	}
	return er
}

func loadApproval(t *testing.T, db *gorm.DB, id uint) models.ExpenseApprovals {
	t.Helper()
	var a models.ExpenseApprovals
	if err := db.First(&a, id).Error; err != nil {
		t.Fatalf("load approval %d: %v", id, err)
	}
	return a
}

// A two-level chain: level 1 approving advances the request without finalising
// it; level 2 approving finalises it.
func TestApprovalChainAdvancesLevelByLevel(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	repo := newExpenseApprovalsRepo(db)

	l2 := b.NewUser(t, db, "Second Approver", "l2@example.com")
	er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500})
	chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID, l2.ID)

	// --- Level 1 approves ---
	err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
		Status: "approved", Comments: testutil.Ptr("looks fine"),
	})
	if err != nil {
		t.Fatalf("level 1 approve: %v", err)
	}

	got := loadExpense(t, db, er.ID)
	if got.Status != "pending" {
		t.Errorf("after level 1: status = %q, want %q", got.Status, "pending")
	}
	if got.CurrentApproverLevel != 2 {
		t.Errorf("after level 1: current_approver_level = %d, want 2", got.CurrentApproverLevel)
	}
	if a := loadApproval(t, db, chain[0].ID); a.Status != "approved" {
		t.Errorf("level 1 approval status = %q, want approved", a.Status)
	} else if a.ApprovalDate == nil {
		t.Error("level 1 approval_date was not stamped")
	}

	// --- Level 2 approves, finalising ---
	err = repo.UpdateExpenseApproval(chain[1].ID, &models.ExpenseApprovals{
		Status: "approved", Comments: testutil.Ptr("approved"),
	})
	if err != nil {
		t.Fatalf("level 2 approve: %v", err)
	}

	got = loadExpense(t, db, er.ID)
	if got.Status != "approved" {
		t.Errorf("after level 2: status = %q, want %q", got.Status, "approved")
	}
	if got.CurrentApproverLevel != 3 {
		t.Errorf("after level 2: current_approver_level = %d, want 3 (walks past the last level)", got.CurrentApproverLevel)
	}
}

// Rejecting at any level finalises the request immediately and pins
// current_approver_level to the rejecting level.
func TestApprovalChainRejectionAtEachLevel(t *testing.T) {
	db := testutil.DB(t)

	tests := []struct {
		name        string
		rejectAt    int // index into the chain
		approveInto int // how many levels approve first
		wantLevel   uint
	}{
		{name: "rejected at level 1", rejectAt: 0, approveInto: 0, wantLevel: 1},
		{name: "rejected at level 2", rejectAt: 1, approveInto: 1, wantLevel: 2},
		{name: "rejected at level 3", rejectAt: 2, approveInto: 2, wantLevel: 3},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b := testutil.SeedBaseline(t, db)
			repo := newExpenseApprovalsRepo(db)

			l2 := b.NewUser(t, db, "L2", "l2@example.com")
			l3 := b.NewUser(t, db, "L3", "l3@example.com")
			er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500})
			chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID, l2.ID, l3.ID)

			for i := 0; i < tc.approveInto; i++ {
				if err := repo.UpdateExpenseApproval(chain[i].ID, &models.ExpenseApprovals{
					Status: "approved", Comments: testutil.Ptr("ok"),
				}); err != nil {
					t.Fatalf("approve level %d: %v", i+1, err)
				}
			}

			if err := repo.UpdateExpenseApproval(chain[tc.rejectAt].ID, &models.ExpenseApprovals{
				Status: "rejected", Comments: testutil.Ptr("not justified"),
			}); err != nil {
				t.Fatalf("reject: %v", err)
			}

			got := loadExpense(t, db, er.ID)
			if got.Status != "rejected" {
				t.Errorf("status = %q, want rejected", got.Status)
			}
			if got.CurrentApproverLevel != tc.wantLevel {
				t.Errorf("current_approver_level = %d, want %d", got.CurrentApproverLevel, tc.wantLevel)
			}
		})
	}
}

// The three concurrency guards that keep two approvers from racing each other.
func TestApprovalChainRefusesOutOfOrderActions(t *testing.T) {
	db := testutil.DB(t)

	t.Run("an approver above the current level cannot act yet", func(t *testing.T) {
		testutil.Reset(t, db)
		b := testutil.SeedBaseline(t, db)
		repo := newExpenseApprovalsRepo(db)

		l2 := b.NewUser(t, db, "L2", "l2@example.com")
		er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500})
		chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID, l2.ID)

		err := repo.UpdateExpenseApproval(chain[1].ID, &models.ExpenseApprovals{
			Status: "approved", Comments: testutil.Ptr("jumping the queue"),
		})
		if err == nil {
			t.Fatal("expected the level-2 approver to be refused while level 1 is outstanding")
		}
		if !strings.Contains(err.Error(), "advanced past your level") {
			t.Errorf("error = %q, want it to mention advancing past the level", err.Error())
		}
	})

	t.Run("an already-processed approval cannot be re-decided", func(t *testing.T) {
		testutil.Reset(t, db)
		b := testutil.SeedBaseline(t, db)
		repo := newExpenseApprovalsRepo(db)

		l2 := b.NewUser(t, db, "L2", "l2@example.com")
		er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500})
		chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID, l2.ID)

		if err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
			Status: "approved", Comments: testutil.Ptr("ok"),
		}); err != nil {
			t.Fatalf("first approve: %v", err)
		}

		err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
			Status: "rejected", Comments: testutil.Ptr("changed my mind"),
		})
		if err == nil {
			t.Fatal("expected a second decision on the same approval to be refused")
		}
	})

	t.Run("a finalised request rejects further decisions", func(t *testing.T) {
		testutil.Reset(t, db)
		b := testutil.SeedBaseline(t, db)
		repo := newExpenseApprovalsRepo(db)

		er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500, Status: "approved"})
		chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID)

		err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
			Status: "approved", Comments: testutil.Ptr("late"),
		})
		if err == nil {
			t.Fatal("expected a decision on an already-finalised request to be refused")
		}
		if !strings.Contains(err.Error(), "already been finalized") {
			t.Errorf("error = %q, want it to mention finalization", err.Error())
		}
	})
}

// Final approval of an expense that fully settles its advance must complete the
// advance too — the multi-use settlement rule.
func TestFinalApprovalCompletesFullySettledAdvance(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	repo := newExpenseApprovalsRepo(db)

	ar := b.NewAdvance(t, db, 1000, "approved")
	er := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 1000, AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(1000.0),
	})
	chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID)

	if err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
		Status: "approved", Comments: testutil.Ptr("ok"),
	}); err != nil {
		t.Fatalf("approve: %v", err)
	}

	if got := loadExpense(t, db, er.ID); got.Status != "approved" {
		t.Fatalf("expense status = %q, want approved", got.Status)
	}

	var gotAR models.AdvanceRequests
	if err := db.First(&gotAR, ar.ID).Error; err != nil {
		t.Fatalf("load advance: %v", err)
	}
	if gotAR.Status != "completed" {
		t.Errorf("advance status = %q, want completed", gotAR.Status)
	}
}

// A partial draw leaves the advance open for the next expense to draw against.
func TestFinalApprovalLeavesPartiallySettledAdvanceApproved(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	repo := newExpenseApprovalsRepo(db)

	ar := b.NewAdvance(t, db, 1000, "approved")
	er := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 400, AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(400.0),
	})
	chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID)

	if err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
		Status: "approved", Comments: testutil.Ptr("ok"),
	}); err != nil {
		t.Fatalf("approve: %v", err)
	}

	var gotAR models.AdvanceRequests
	if err := db.First(&gotAR, ar.ID).Error; err != nil {
		t.Fatalf("load advance: %v", err)
	}
	if gotAR.Status != "approved" {
		t.Errorf("advance status = %q, want approved (600 still outstanding)", gotAR.Status)
	}
}

func TestRejectionWithoutCommentIsAccepted(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	repo := newExpenseApprovalsRepo(db)

	er := b.NewExpense(t, db, testutil.ExpenseOpts{Amount: 500})
	chain := b.NewApprovalChain(t, db, er.ID, b.Approver.ID)

	if err := repo.UpdateExpenseApproval(chain[0].ID, &models.ExpenseApprovals{
		Status: "rejected", Comments: nil,
	}); err != nil {
		t.Fatalf("rejecting without a comment should succeed, got %v", err)
	}

	if got := loadExpense(t, db, er.ID); got.Status != "rejected" {
		t.Errorf("expense status = %q, want rejected", got.Status)
	}
	if got := loadApproval(t, db, chain[0].ID); got.Status != "rejected" {
		t.Errorf("approval status = %q, want rejected", got.Status)
	}
}

func TestAdvanceRejectionWithoutCommentSucceeds(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	repo := repositories.NewAdvanceApprovalsRepo(db, nil)

	ar := b.NewAdvance(t, db, 1000, "pending")
	chain := b.NewAdvanceApprovalChain(t, db, ar.ID, b.Approver.ID)

	if err := repo.UpdateAdvanceApproval(chain[0].ID, &models.AdvanceApprovals{
		Status: "rejected", Comments: nil,
	}); err != nil {
		t.Fatalf("reject advance without comment: %v", err)
	}

	var got models.AdvanceRequests
	if err := db.First(&got, ar.ID).Error; err != nil {
		t.Fatalf("load advance: %v", err)
	}
	if got.Status != "rejected" {
		t.Errorf("advance status = %q, want rejected", got.Status)
	}
}
