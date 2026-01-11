package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"gorm.io/gorm"
)

type VaultRepository struct {
	db *gorm.DB
}

func NewVaultRepository(db *gorm.DB) *VaultRepository {
	return &VaultRepository{db: db}
}

func (r *VaultRepository) Create(ctx context.Context, vault *models.Vault) error {
	return r.db.WithContext(ctx).Create(vault).Error
}

func (r *VaultRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Vault, error) {
	var vault models.Vault
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&vault).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &vault, err
}

func (r *VaultRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]models.Vault, error) {
	var vaults []models.Vault
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC").Find(&vaults).Error
	return vaults, err
}

func (r *VaultRepository) Update(ctx context.Context, vault *models.Vault) error {
	return r.db.WithContext(ctx).Save(vault).Error
}

func (r *VaultRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Vault{}, id).Error
}

func (r *VaultRepository) Search(ctx context.Context, userID uuid.UUID, searchTerm string) ([]models.Vault, error) {
	var vaults []models.Vault
	searchPattern := "%" + searchTerm + "%"
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Where("title ILIKE ? OR website ILIKE ? OR username ILIKE ?", searchPattern, searchPattern, searchPattern).
		Order("created_at DESC").
		Find(&vaults).Error
	return vaults, err
}
