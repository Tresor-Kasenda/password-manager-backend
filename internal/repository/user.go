package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/tresor/password-manager/internal/models"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
        INSERT INTO users (
            id, email, master_password_hash, salt, 
            public_key, private_key, two_factor_enabled,
            created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
    `

	_, err := r.db.ExecContext(
		ctx, query,
		user.ID,
		user.Email,
		user.MasterPasswordHash,
		user.Salt,
		user.PublicKey,
		user.PrivateKey,
		user.TwoFactorEnabled,
		user.CreatedAt,
		user.UpdatedAt,
	)

	return err
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	query := `
        SELECT id, email, master_password_hash, salt, 
               public_key, private_key, two_factor_enabled,
               two_factor_secret, backup_codes,
               created_at, updated_at
        FROM users
        WHERE id = $1
    `

	err := r.db.GetContext(ctx, &user, query, id)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &user, err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	query := `
        SELECT id, email, master_password_hash, salt, 
               public_key, private_key, two_factor_enabled,
               two_factor_secret, backup_codes,
               created_at, updated_at
        FROM users
        WHERE email = $1
    `

	err := r.db.GetContext(ctx, &user, query, email)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &user, err
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
        UPDATE users
        SET email = $2,
            master_password_hash = $3,
            salt = $4,
            public_key = $5,
            private_key = $6,
            two_factor_enabled = $7,
            two_factor_secret = $8,
            backup_codes = $9,
            updated_at = $10
        WHERE id = $1
    `

	_, err := r.db.ExecContext(
		ctx, query,
		user.ID,
		user.Email,
		user.MasterPasswordHash,
		user.Salt,
		user.PublicKey,
		user.PrivateKey,
		user.TwoFactorEnabled,
		user.TwoFactorSecret,
		pq.Array(user.BackupCodes),
		user.UpdatedAt,
	)

	return err
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
