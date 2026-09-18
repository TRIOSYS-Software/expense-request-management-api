package repositories_test

import (
	"strings"
	"testing"

	"shwetaik-expense-management-api/models"
	"shwetaik-expense-management-api/repositories"
	"shwetaik-expense-management-api/services"
	"shwetaik-expense-management-api/storage"
	"shwetaik-expense-management-api/tests/testutil"

	"gorm.io/gorm"
)

func TestValidateAdvanceLinkWithoutAdvanceClearsAmounts(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)

	req := &models.ExpenseRequests{
		UserID:            b.User.ID,
		Amount:            250,
		AdvanceRequestID:  nil,
		AdvanceUsedAmount: testutil.Ptr(100.0),
		ReturnedAmount:    testutil.Ptr(50.0),
	}

	if err := services.ValidateAdvanceLink(repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir())), req, nil); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.AdvanceUsedAmount != nil {
		t.Errorf("AdvanceUsedAmount = %v, want nil", *req.AdvanceUsedAmount)
	}
	if req.ReturnedAmount != nil {
		t.Errorf("ReturnedAmount = %v, want nil", *req.ReturnedAmount)
	}
}

func TestValidateAdvanceLinkRejections(t *testing.T) {
	db := testutil.DB(t)

	tests := []struct {
		name    string
		build   func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests
		wantErr string
	}{
		{
			name: "negative amount",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				return &models.ExpenseRequests{UserID: b.User.ID, Amount: -1}
			},
			wantErr: "Expense amount cannot be negative",
		},
		{
			name: "linked advance with no used amount",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: nil,
				}
			},
			wantErr: "Advance used amount is required and must be greater than zero",
		},
		{
			name: "linked advance with zero used amount",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(0.0),
				}
			},
			wantErr: "Advance used amount is required and must be greater than zero",
		},
		{
			name: "returned amount of zero is rejected as explicitly provided",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID:  &ar.ID,
					AdvanceUsedAmount: testutil.Ptr(500.0),
					ReturnedAmount:    testutil.Ptr(0.0),
				}
			},
			wantErr: "Returned amount, when provided, must be greater than zero",
		},
		{
			name: "advance does not exist",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				missing := uint(999999)
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID: &missing, AdvanceUsedAmount: testutil.Ptr(50.0),
				}
			},
			wantErr: "Linked advance request not found",
		},
		{
			name: "advance belongs to another user",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				other := b.NewUser(t, db, "Other", "other@example.com")
				return &models.ExpenseRequests{
					UserID: other.ID, Amount: 100,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(50.0),
				}
			},
			wantErr: "You may only link an advance request created by yourself",
		},
		{
			name: "advance is still pending",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "pending")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(50.0),
				}
			},
			wantErr: "Only an approved advance request may be linked",
		},
		{
			name: "advance was rejected",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "rejected")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 100,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(50.0),
				}
			},
			wantErr: "Only an approved advance request may be linked",
		},
		{
			name: "draw exceeds the advance total",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 2000,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(1500.0),
				}
			},
			wantErr: "exceeds the advance request's remaining balance",
		},
		{
			name: "draw exceeds what earlier expenses left behind",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				b.NewExpense(t, db, testutil.ExpenseOpts{
					Amount: 800, Status: "approved",
					AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(800.0),
				})
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 300,
					AdvanceRequestID: &ar.ID, AdvanceUsedAmount: testutil.Ptr(300.0),
				}
			},
			wantErr: "exceeds the advance request's remaining balance",
		},
		{
			name: "returning more than went unspent",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				// Drew 500, spent 400 — only 100 is genuinely left over.
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 400,
					AdvanceRequestID:  &ar.ID,
					AdvanceUsedAmount: testutil.Ptr(500.0),
					ReturnedAmount:    testutil.Ptr(200.0),
				}
			},
			wantErr: "cannot exceed the unused portion of the advance used",
		},
		{
			name: "returning anything when the whole draw was spent",
			build: func(t *testing.T, db *gorm.DB, b testutil.Baseline) *models.ExpenseRequests {
				ar := b.NewAdvance(t, db, 1000, "approved")
				return &models.ExpenseRequests{
					UserID: b.User.ID, Amount: 500,
					AdvanceRequestID:  &ar.ID,
					AdvanceUsedAmount: testutil.Ptr(500.0),
					ReturnedAmount:    testutil.Ptr(10.0),
				}
			},
			wantErr: "cannot exceed the unused portion of the advance used",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)

			b := testutil.SeedBaseline(t, db)

			req := tc.build(t, db, b)
			err := services.ValidateAdvanceLink(repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir())), req, nil)
			if err == nil {
				t.Fatalf("expected error containing %q, got nil", tc.wantErr)
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestValidateAdvanceLinkAccepted(t *testing.T) {
	db := testutil.DB(t)

	tests := []struct {
		name        string
		expenseAmt  float64
		advanceUsed float64
		returned    *float64
	}{
		{name: "spending the entire draw", expenseAmt: 500, advanceUsed: 500},
		{name: "spending less and keeping the rest on the advance", expenseAmt: 300, advanceUsed: 500},
		{name: "returning exactly the unspent portion", expenseAmt: 300, advanceUsed: 500, returned: testutil.Ptr(200.0)},
		{name: "returning part of the unspent portion", expenseAmt: 300, advanceUsed: 500, returned: testutil.Ptr(120.0)},
		{name: "drawing the full advance", expenseAmt: 1000, advanceUsed: 1000},
		{name: "a zero-amount expense against a draw", expenseAmt: 0, advanceUsed: 500, returned: testutil.Ptr(500.0)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			testutil.Reset(t, db)
			b := testutil.SeedBaseline(t, db)
			ar := b.NewAdvance(t, db, 1000, "approved")

			req := &models.ExpenseRequests{
				UserID: b.User.ID, Amount: tc.expenseAmt,
				AdvanceRequestID:  &ar.ID,
				AdvanceUsedAmount: testutil.Ptr(tc.advanceUsed),
				ReturnedAmount:    tc.returned,
			}
			if err := services.ValidateAdvanceLink(repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir())), req, nil); err != nil {
				t.Fatalf("expected the link to be accepted, got %v", err)
			}
		})
	}
}

func TestValidateAdvanceLinkExcludesTheRowBeingEdited(t *testing.T) {
	db := testutil.DB(t)
	b := testutil.SeedBaseline(t, db)
	ar := b.NewAdvance(t, db, 1000, "approved")

	existing := b.NewExpense(t, db, testutil.ExpenseOpts{
		Amount: 1000, Status: "approved",
		AdvanceRequestID: &ar.ID, AdvanceUsed: testutil.Ptr(1000.0),
	})

	edited := &models.ExpenseRequests{
		UserID: b.User.ID, Amount: 900,
		AdvanceRequestID:  &ar.ID,
		AdvanceUsedAmount: testutil.Ptr(900.0),
	}

	if err := services.ValidateAdvanceLink(repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir())), edited, nil); err == nil {
		t.Fatal("expected the un-excluded validation to fail on a fully drawn advance")
	}

	if err := services.ValidateAdvanceLink(repositories.NewExpenseRequestsRepo(db, nil, storage.NewDiskStore(t.TempDir())), edited, &existing.ID); err != nil {
		t.Fatalf("expected the excluded validation to pass, got %v", err)
	}
}
