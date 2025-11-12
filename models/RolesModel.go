package models

type Roles struct {
	ID          uint          `json:"id" gorm:"primaryKey;autoIncrement;unique"`
	Name        string        `json:"name" gorm:"not null"`
	Description string        `json:"description" gorm:"type:text"`
	IsAdmin     bool          `json:"is_admin" gorm:"default:false;not null"`
	Permissions []Permissions `json:"permissions" gorm:"many2many:roles_permissions;constraint:OnDelete:CASCADE;"`
	UserCount   int64         `json:"user_count" gorm:"-"`
}
