package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex"`
	Username string `gorm:"uniqueIndex"`
	Password  string
	Currency  string
	Balance   int
	Character string
	NewUser   bool
}
