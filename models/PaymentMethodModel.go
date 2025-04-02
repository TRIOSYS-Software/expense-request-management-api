package models

type PaymentMethod struct {
	CODE         string `gorm:"column:CODE;primaryKey" json:"code,omitempty"`
	JOURNAL      string `gorm:"column:JOURNAL" json:"journal,omitempty"`
	CURRENCYCODE string `gorm:"column:CURRENCYCODE" json:"currency_code,omitempty"`
	DESCRIPTION  string `gorm:"column:DESCRIPTION" json:"description,omitempty"`
}
