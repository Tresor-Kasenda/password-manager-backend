package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type User struct {
	ID                 uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	Email              string         `gorm:"uniqueIndex;not null" json:"email"`
	MasterPasswordHash string         `gorm:"not null" json:"-"`
	Salt               string         `gorm:"not null" json:"-"`
	PublicKey          string         `json:"public_key,omitempty"`
	PrivateKey         string         `json:"-"`
	TwoFactorEnabled   bool           `gorm:"default:false" json:"two_factor_enabled"`
	TwoFactorSecret    *string        `json:"-"`
	BackupCodes        pq.StringArray `gorm:"type:text[]" json:"-"`
	CreatedAt          time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt          time.Time      `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	Vaults            []Vault         `gorm:"foreignKey:UserID" json:"-"`
	SharedPasswords   []SharedPassword `gorm:"foreignKey:OwnerID" json:"-"`
}

// TableName specifies the table name for GORM
func (User) TableName() string {
	return "users"
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
