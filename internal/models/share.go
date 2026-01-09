package models

import (
	"time"

	"github.com/google/uuid"
)

type SharedPassword struct {
	ID                uuid.UUID  `db:"id" json:"id"`
	VaultID           uuid.UUID  `db:"vault_id" json:"vault_id"`
	OwnerID           uuid.UUID  `db:"owner_id" json:"owner_id"`
	RecipientID       *uuid.UUID `db:"recipient_id" json:"recipient_id"`
	RecipientEmail    string     `db:"recipient_email" json:"recipient_email"`
	EncryptedData     string     `db:"encrypted_data" json:"encrypted_data"`
	ShareToken        string     `db:"share_token" json:"share_token"`
	ExpiresAt         *time.Time `db:"expires_at" json:"expires_at"`
	MaxViews          *int       `db:"max_views" json:"max_views"`
	ViewCount         int        `db:"view_count" json:"view_count"`
	RequirePassword   bool       `db:"require_password" json:"require_password"`
	SharePasswordHash *string    `db:"share_password_hash" json:"-"`
	CanView           bool       `db:"can_view" json:"can_view"`
	CanCopy           bool       `db:"can_copy" json:"can_copy"`
	CanEdit           bool       `db:"can_edit" json:"can_edit"`
	Revoked           bool       `db:"revoked" json:"revoked"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	LastAccessed      *time.Time `db:"last_accessed" json:"last_accessed"`
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
