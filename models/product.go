package models

import "time"

// Product represents a catalog item owned by a user.
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name" validate:"required"`
	Description string    `json:"description"`
	Price       float64   `gorm:"not null" json:"price" validate:"required,gt=0"`
	SKU         string    `gorm:"uniqueIndex" json:"sku"`
	UserID      uint      `json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
