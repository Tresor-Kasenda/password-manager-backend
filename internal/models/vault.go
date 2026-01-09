package models

import (
	"time"

	"github.com/google/uuid"
)

type Vault struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	UserID         uuid.UUID  `db:"user_id" json:"user_id"`
	Title          string     `db:"title" json:"title"`
	Website        *string    `db:"website" json:"website"`
	Username       *string    `db:"username" json:"username"`
	EncryptedData  string     `db:"encrypted_data" json:"encrypted_data"`
	EncryptionSalt string     `db:"encryption_salt" json:"encryption_salt"`
	Nonce          string     `db:"nonce" json:"nonce"`
	Folder         *string    `db:"folder" json:"folder"`
	Favorite       bool       `db:"favorite" json:"favorite"`
	LastUsed       *time.Time `db:"last_used" json:"last_used"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updated_at"`
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
