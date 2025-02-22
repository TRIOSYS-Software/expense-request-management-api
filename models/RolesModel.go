package models

type Roles struct {
	ID   uint   `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name string `json:"name" gorm:"not null"`
}
