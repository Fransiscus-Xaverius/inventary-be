package models

import (
	"time"
)

// Newsletter represents a newsletter subscription record in the database
type Newsletter struct {
	ID        int        `json:"id" form:"id"`
	Email     string     `json:"email" form:"email" binding:"required,email"`
	Whatsapp  string     `json:"whatsapp" form:"whatsapp" binding:"required"`
	Message   string     `json:"message" form:"message"`
	CreatedAt time.Time  `json:"created_at" form:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" form:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" form:"deleted_at,omitempty"`
}
