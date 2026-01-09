package models

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID                 uuid.UUID `db:"id" json:"id"`
	Email              string    `db:"email" json:"email"`
	MasterPasswordHash string    `db:"master_password_hash" json:"-"`
	Salt               string    `db:"salt" json:"-"`
	PublicKey          string    `db:"public_key" json:"public_key,omitempty"`
	PrivateKey         string    `db:"private_key" json:"-"`
	TwoFactorEnabled   bool      `db:"two_factor_enabled" json:"two_factor_enabled"`
	TwoFactorSecret    *string   `db:"two_factor_secret" json:"-"`
	BackupCodes        []string  `db:"backup_codes" json:"-"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

type LoginRequest struct {
	Email          string `json:"email" binding:"required,email"`
	MasterPassword string `json:"master_password" binding:"required"`
}

type RegisterRequest struct {
	Email          string `json:"email" binding:"required,email"`
	MasterPassword string `json:"master_password" binding:"required,min=8"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	User        User   `json:"user"`
}
