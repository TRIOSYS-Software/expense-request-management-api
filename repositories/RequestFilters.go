package repositories

import (
	"strconv"

	"gorm.io/gorm"
)

const (
	searchUsersAlias    = "search_users"
	searchProjectsAlias = "search_projects"
)

func applyDateRangeFilter(db *gorm.DB, table, startDate, endDate string) *gorm.DB {
	switch {
	case startDate != "" && endDate != "":
		return db.Where("DATE("+table+".date_submitted) BETWEEN ? AND ?", startDate, endDate)
	case startDate != "":
		return db.Where("DATE("+table+".date_submitted) >= ?", startDate)
	case endDate != "":
		return db.Where("DATE("+table+".date_submitted) <= ?", endDate)
	default:
		return db
	}
}

func applySearchFilter(db *gorm.DB, table, search string) *gorm.DB {
	if search == "" {
		return db
	}

	db = db.
		Joins("LEFT JOIN users " + searchUsersAlias + " ON " + searchUsersAlias + ".id = " + table + ".user_id").
		Joins("LEFT JOIN projects " + searchProjectsAlias + " ON " + searchProjectsAlias + ".CODE = " + table + ".project")

	pattern := "%" + search + "%"
	textColumns := table + ".description LIKE ? OR " +
		searchUsersAlias + ".name LIKE ? OR " +
		searchProjectsAlias + ".CODE LIKE ? OR " +
		searchProjectsAlias + ".DESCRIPTION LIKE ?"

	if id, err := strconv.Atoi(search); err == nil {
		return db.Where("("+table+".id = ? OR "+textColumns+")", id, pattern, pattern, pattern, pattern)
	}
	return db.Where("("+textColumns+")", pattern, pattern, pattern, pattern)
}

func applyAmountRangeFilter(db *gorm.DB, table string, minAmount, maxAmount *float64) *gorm.DB {
	if minAmount != nil {
		db = db.Where(table+".amount >= ?", *minAmount)
	}
	if maxAmount != nil {
		db = db.Where(table+".amount <= ?", *maxAmount)
	}
	return db
}

type requestListFilter struct {
	StartDate string
	EndDate   string
	Search    string
	MinAmount *float64
	MaxAmount *float64
}

func applyRequestListFilters(db *gorm.DB, table string, f *requestListFilter) *gorm.DB {
	if f == nil {
		return db
	}
	db = applyDateRangeFilter(db, table, f.StartDate, f.EndDate)
	db = applySearchFilter(db, table, f.Search)
	return applyAmountRangeFilter(db, table, f.MinAmount, f.MaxAmount)
}

func applySummaryFilters(db *gorm.DB, table string, filters map[string]any) *gorm.DB {
	startDate, _ := filters["start_date"].(string)
	endDate, _ := filters["end_date"].(string)
	if startDate != "" && endDate != "" {
		db = applyDateRangeFilter(db, table, startDate, endDate)
	}

	if search, ok := filters["search"].(string); ok {
		db = applySearchFilter(db, table, search)
	}
	return db
}
