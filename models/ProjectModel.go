package models

type Project struct {
	CODE         string   `gorm:"column:CODE;primaryKey" json:"code,omitempty"`
	DESCRIPTION  *string  `gorm:"column:DESCRIPTION" json:"description,omitempty"`
	DESCRIPTION2 *string  `gorm:"column:DESCRIPTION2" json:"description2,omitempty"`
	PROJECTVALUE *float64 `gorm:"column:PROJECTVALUE;default:0" json:"project_value,omitempty"`
	PROJECTCOST  *float64 `gorm:"column:PROJECTCOST;default:0" json:"project_cost,omitempty"`
	ATTACHMENTS  *[]byte  `gorm:"column:ATTACHMENTS" json:"attachments,omitempty"`
	ISACTIVE     *bool    `gorm:"column:ISACTIVE" json:"is_active,omitempty"`
	Users        []Users  `json:"users,omitempty" gorm:"many2many:users_projects;"`
}
