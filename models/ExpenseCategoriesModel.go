package models

type ExpenseCategories struct {
	ID   uint   `json:"id,omitempty" gorm:"primaryKey;autoIncrement;unique"`
	Name string `json:"name,omitempty" gorm:"not null"`
}
