package models

type ExpenseCategories struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name string `json:"name" gorm:"not null"`
}
