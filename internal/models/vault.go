package models

import (
	"time"

	"github.com/google/uuid"
)

type Vault struct {
	ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"user_id"`
	Title          string     `gorm:"not null" json:"title"`
	Website        *string    `json:"website,omitempty"`
	Username       *string    `json:"username,omitempty"`
	EncryptedData  string     `gorm:"not null" json:"-"`
	EncryptionSalt string     `gorm:"not null" json:"-"`
	Nonce          string     `gorm:"not null" json:"-"`
	Folder         *string    `json:"folder,omitempty"`
	Favorite       bool       `gorm:"default:false" json:"favorite"`
	LastUsed       *time.Time `json:"last_used,omitempty"`
	CreatedAt      time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relations
	User User `gorm:"foreignKey:UserID" json:"-"`
}

// TableName specifies the table name for GORM
func (Vault) TableName() string {
	return "vaults"
}

type CreateVaultRequest struct {
	Title          string  `json:"title" binding:"required"`
	Website        *string `json:"website"`
	Username       *string `json:"username"`
	Password       string  `json:"password" binding:"required"`
	Notes          *string `json:"notes"`
	Folder         *string `json:"folder"`
	MasterPassword string  `json:"master_password" binding:"required"`
}

type VaultResponse struct {
	ID        uuid.UUID `json:"id"`
	Title     string    `json:"title"`
	Website   *string   `json:"website"`
	Username  *string   `json:"username"`
	Folder    *string   `json:"folder"`
	Favorite  bool      `json:"favorite"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DecryptedVaultData struct {
	Password string  `json:"password"`
	Notes    *string `json:"notes"`
}
