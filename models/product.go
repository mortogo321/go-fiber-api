package models

import "time"

// Product represents a product owned by a user.
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	SKU         string    `gorm:"uniqueIndex;size:100;not null" json:"sku"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	User        User      `gorm:"foreignKey:UserID" json:"-"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}
