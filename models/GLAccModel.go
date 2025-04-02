package models

type GLAcc struct {
	DOCKEY         int    `gorm:"column:DOCKEY;primaryKey" json:"dockey,omitempty"`
	PARENT         int    `gorm:"column:PARENT" json:"parent,omitempty"`
	CODE           string `gorm:"column:CODE" json:"code,omitempty"`
	DESCRIPTION    string `gorm:"column:DESCRIPTION" json:"description,omitempty"`
	DESCRIPTION2   string `gorm:"column:DESCRIPTION2" json:"description2,omitempty"`
	ACCTYPE        string `gorm:"column:ACCTYPE" json:"acctype,omitempty"`
	SPECIALACCTYPE string `gorm:"column:SPECIALACCTYPE" json:"special_acctype,omitempty"`
	TAX            string `gorm:"column:TAX" json:"tax,omitempty"`
	CASHFLOWTYPE   int    `gorm:"column:CASHFLOWTYPE" json:"cashflow_type,omitempty"`
	SIC            string `gorm:"column:SIC" json:"sic,omitempty"`
}
