package models

type Permissions struct {
	ID         uint   `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name       string `json:"name" gorm:"not null"`
	ActionName string `json:"action_name" gorm:"not null"`
	Entity     string `json:"entity" gorm:"not null;index:idx_entity_action,unique"`
	Action     string `json:"action" gorm:"not null;index:idx_entity_action,unique"`
}
