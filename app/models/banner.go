package models

import (
	"time"
)

// Banner represents a banner record in the database
type Banner struct {
	ID          int        `json:"id" form:"id"`
	Title       string     `json:"title" form:"title"`
	Description string     `json:"description" form:"description"`
	CtaText     string     `json:"cta_text" form:"cta_text"`
	CtaLink     string     `json:"cta_link" form:"cta_link"`
	ImageUrl    string     `json:"image_url" form:"image_url"`
	OrderIndex  int        `json:"order_index" form:"order_index"`
	IsActive    bool       `json:"is_active" form:"is_active"`
	CreatedAt   time.Time  `json:"created_at" form:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" form:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" form:"deleted_at,omitempty"`
}
