package models

import "time"

// User represents an application user with authentication credentials.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	Password  string    `gorm:"not null" json:"-"`
	Name      string    `gorm:"not null" json:"name"`
	Role      string    `gorm:"default:user" json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
