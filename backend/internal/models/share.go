package models

import (
	"time"

	"github.com/google/uuid"
)

type SharedPassword struct {
	ID                uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	VaultID           uuid.UUID  `gorm:"type:uuid;not null" json:"vault_id"`
	OwnerID           uuid.UUID  `gorm:"type:uuid;not null" json:"owner_id"`
	RecipientID       *uuid.UUID `gorm:"type:uuid" json:"recipient_id"`
	RecipientEmail    string     `gorm:"not null" json:"recipient_email"`
	EncryptedData     string     `gorm:"not null" json:"encrypted_data"`
	ShareToken        string     `gorm:"uniqueIndex;not null" json:"share_token"`
	ExpiresAt         *time.Time `json:"expires_at,omitempty"`
	MaxViews          *int       `json:"max_views,omitempty"`
	ViewCount         int        `gorm:"default:0" json:"view_count"`
	RequirePassword   bool       `gorm:"default:false" json:"require_password"`
	SharePasswordHash *string    `json:"-"`
	CanView           bool       `gorm:"default:true" json:"can_view"`
	CanCopy           bool       `gorm:"default:true" json:"can_copy"`
	CanEdit           bool       `gorm:"default:false" json:"can_edit"`
	Revoked           bool       `gorm:"default:false" json:"revoked"`
	CreatedAt         time.Time  `gorm:"autoCreateTime" json:"created_at"`
	LastAccessed      *time.Time `json:"last_accessed,omitempty"`

	// Relations
	Vault     Vault  `gorm:"foreignKey:VaultID" json:"-"`
	Owner     User   `gorm:"foreignKey:OwnerID" json:"-"`
	Recipient *User  `gorm:"foreignKey:RecipientID" json:"-"`
}

// TableName specifies the table name for GORM
func (SharedPassword) TableName() string {
	return "shared_passwords"
}

type SharePasswordRequest struct {
	VaultID         string  `json:"vault_id" binding:"required"`
	RecipientEmail  string  `json:"recipient_email" binding:"required,email"`
	ExpiresInHours  *int    `json:"expires_in_hours"`
	MaxViews        *int    `json:"max_views"`
	RequirePassword bool    `json:"require_password"`
	SharePassword   *string `json:"share_password"`
	CanEdit         bool    `json:"can_edit"`
}

type SharePasswordResponse struct {
	ShareToken string     `json:"share_token"`
	ShareURL   string     `json:"share_url"`
	ExpiresAt  *time.Time `json:"expires_at"`
}
