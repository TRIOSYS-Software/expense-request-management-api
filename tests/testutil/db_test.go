package testutil

import (
	"testing"

	"shwetaik-expense-management-api/models"
)

func TestHarnessMigratesAndResets(t *testing.T) {
	db := DB(t)

	if err := db.Create(&models.Departments{Name: "Engineering"}).Error; err != nil {
		t.Fatalf("insert department: %v", err)
	}

	var count int64
	if err := db.Model(&models.Departments{}).Count(&count).Error; err != nil {
		t.Fatalf("count departments: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 department after insert, got %d", count)
	}

	Reset(t, db)

	if err := db.Model(&models.Departments{}).Count(&count).Error; err != nil {
		t.Fatalf("count departments after reset: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 departments after Reset, got %d", count)
	}
}

func TestHarnessIsolatesBetweenTests(t *testing.T) {
	db := DB(t)

	var count int64
	if err := db.Model(&models.Departments{}).Count(&count).Error; err != nil {
		t.Fatalf("count departments: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected a clean database at test start, got %d departments", count)
	}
}
