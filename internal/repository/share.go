package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/tresor/password-manager/internal/models"
)

type ShareRepository struct {
	db *sqlx.DB
}

func NewShareRepository(db *sqlx.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Create(ctx context.Context, share *models.SharedPassword) error {
	query := `
        INSERT INTO shared_passwords (
            id, vault_id, owner_id, recipient_id, recipient_email,
            encrypted_data, share_token, expires_at, max_views,
            view_count, require_password, share_password_hash,
            can_view, can_copy, can_edit, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
    `

	_, err := r.db.ExecContext(
		ctx, query,
		share.ID,
		share.VaultID,
		share.OwnerID,
		share.RecipientID,
		share.RecipientEmail,
		share.EncryptedData,
		share.ShareToken,
		share.ExpiresAt,
		share.MaxViews,
		share.ViewCount,
		share.RequirePassword,
		share.SharePasswordHash,
		share.CanView,
		share.CanCopy,
		share.CanEdit,
		share.CreatedAt,
	)

	return err
}

func (r *ShareRepository) GetByToken(ctx context.Context, token string) (*models.SharedPassword, error) {
	var share models.SharedPassword
	query := `
        SELECT id, vault_id, owner_id, recipient_id, recipient_email,
               encrypted_data, share_token, expires_at, max_views,
               view_count, require_password, share_password_hash,
               can_view, can_copy, can_edit, revoked,
               created_at, last_accessed
        FROM shared_passwords
        WHERE share_token = $1
    `

	err := r.db.GetContext(ctx, &share, query, token)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &share, err
}

func (r *ShareRepository) GetSentShares(ctx context.Context, ownerID uuid.UUID) ([]models.SharedPassword, error) {
	var shares []models.SharedPassword
	query := `
        SELECT id, vault_id, owner_id, recipient_id, recipient_email,
               encrypted_data, share_token, expires_at, max_views,
               view_count, require_password, share_password_hash,
               can_view, can_copy, can_edit, revoked,
               created_at, last_accessed
        FROM shared_passwords
        WHERE owner_id = $1
        ORDER BY created_at DESC
    `

	err := r.db.SelectContext(ctx, &shares, query, ownerID)
	return shares, err
}

func (r *ShareRepository) GetReceivedShares(ctx context.Context, recipientID uuid.UUID) ([]models.SharedPassword, error) {
	var shares []models.SharedPassword
	query := `
        SELECT id, vault_id, owner_id, recipient_id, recipient_email,
               encrypted_data, share_token, expires_at, max_views,
               view_count, require_password, share_password_hash,
               can_view, can_copy, can_edit, revoked,
               created_at, last_accessed
        FROM shared_passwords
        WHERE recipient_id = $1
        ORDER BY created_at DESC
    `

	err := r.db.SelectContext(ctx, &shares, query, recipientID)
	return shares, err
}

func (r *ShareRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	query := `
        UPDATE shared_passwords
        SET view_count = view_count + 1,
            last_accessed = NOW()
        WHERE id = $1
    `

	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *ShareRepository) Revoke(ctx context.Context, token string, ownerID uuid.UUID) error {
	query := `
        UPDATE shared_passwords
        SET revoked = true
        WHERE share_token = $1 AND owner_id = $2
    `

	_, err := r.db.ExecContext(ctx, query, token, ownerID)
	return err
}
