package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/tresor/password-manager/internal/models"
)

type VaultRepository struct {
	db *sqlx.DB
}

func NewVaultRepository(db *sqlx.DB) *VaultRepository {
	return &VaultRepository{db: db}
}

func (r *VaultRepository) Create(ctx context.Context, vault *models.Vault) error {
	query := `
        INSERT INTO vaults (
            id, user_id, title, website, username,
            encrypted_data, encryption_salt, nonce,
            folder, favorite, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
    `

	_, err := r.db.ExecContext(
		ctx, query,
		vault.ID,
		vault.UserID,
		vault.Title,
		vault.Website,
		vault.Username,
		vault.EncryptedData,
		vault.EncryptionSalt,
		vault.Nonce,
		vault.Folder,
		vault.Favorite,
		vault.CreatedAt,
		vault.UpdatedAt,
	)

	return err
}

func (r *VaultRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vault, error) {
	var vault models.Vault
	query := `
        SELECT id, user_id, title, website, username,
               encrypted_data, encryption_salt, nonce,
               folder, favorite, last_used,
               created_at, updated_at
        FROM vaults
        WHERE id = $1
    `

	err := r.db.GetContext(ctx, &vault, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &vault, err
}

func (r *VaultRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Vault, error) {
	var vaults []models.Vault
	query := `
        SELECT id, user_id, title, website, username,
               encrypted_data, encryption_salt, nonce,
               folder, favorite, last_used,
               created_at, updated_at
        FROM vaults
        WHERE user_id = $1
        ORDER BY created_at DESC
    `

	err := r.db.SelectContext(ctx, &vaults, query, userID)
	return vaults, err
}

func (r *VaultRepository) Update(ctx context.Context, vault *models.Vault) error {
	query := `
        UPDATE vaults
        SET title = $2,
            website = $3,
            username = $4,
            encrypted_data = $5,
            encryption_salt = $6,
            nonce = $7,
            folder = $8,
            favorite = $9,
            last_used = $10,
            updated_at = $11
        WHERE id = $1
    `

	_, err := r.db.ExecContext(
		ctx, query,
		vault.ID,
		vault.Title,
		vault.Website,
		vault.Username,
		vault.EncryptedData,
		vault.EncryptionSalt,
		vault.Nonce,
		vault.Folder,
		vault.Favorite,
		vault.LastUsed,
		vault.UpdatedAt,
	)

	return err
}

func (r *VaultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM vaults WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *VaultRepository) Search(ctx context.Context, userID uuid.UUID, searchTerm string) ([]models.Vault, error) {
	var vaults []models.Vault
	query := `
        SELECT id, user_id, title, website, username,
               encrypted_data, encryption_salt, nonce,
               folder, favorite, last_used,
               created_at, updated_at
        FROM vaults
        WHERE user_id = $1
          AND (
              title ILIKE $2
              OR website ILIKE $2
              OR username ILIKE $2
          )
        ORDER BY created_at DESC
    `

	searchPattern := "%" + searchTerm + "%"
	err := r.db.SelectContext(ctx, &vaults, query, userID, searchPattern)
	return vaults, err
}
