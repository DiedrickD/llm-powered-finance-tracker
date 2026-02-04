package models

import "gorm.io/gorm"

/*
type Category struct {
	gorm.Model
	Name string `gorm:"unique"`
}
*/

// TODO: Make category unique, flow -> each person can create their own category and
// LLM will assign based on the user category

type Transaction struct {
	gorm.Model
	Amount      int    `json:"amount"`
	Category    string `json:"category"`
	UserID      uint   `json:"user_id"`
	Description string `json:"description"`
}
