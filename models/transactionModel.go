package models

import "gorm.io/gorm"

type Category struct {
	gorm.Model
	Name     string `gorm:"not null;uniqueIndex:idx_user_category"`
	UserID   uint   `gorm:"uniqueIndex:idx_user_category"`
	ParentID *uint
	Parent   *Category `gorm:"foreignkey:ParentID"`
}

type Transaction struct {
	gorm.Model
	Amount      int        `json:"amount"`
	UserID      uint       `json:"user_id"`
	Description string     `json:"description"`
	Categories  []Category `gorm:"many2many:transaction_categories;"`
}
