package models

type Departments struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name string `json:"name" gorm:"not null"`
}
