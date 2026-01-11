package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"gorm.io/gorm"
)

type ShareRepository struct {
	db *gorm.DB
}

func NewShareRepository(db *gorm.DB) *ShareRepository {
	return &ShareRepository{db: db}
}

func (r *ShareRepository) Create(ctx context.Context, share *models.SharedPassword) error {
	return r.db.WithContext(ctx).Create(share).Error
}

func (r *ShareRepository) GetByToken(ctx context.Context, token string) (*models.SharedPassword, error) {
	var share models.SharedPassword
	err := r.db.WithContext(ctx).Where("share_token = ?", token).First(&share).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &share, err
}

func (r *ShareRepository) GetSentShares(ctx context.Context, ownerID uuid.UUID) ([]models.SharedPassword, error) {
	var shares []models.SharedPassword
	err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

func (r *ShareRepository) GetReceivedShares(ctx context.Context, recipientID uuid.UUID) ([]models.SharedPassword, error) {
	var shares []models.SharedPassword
	err := r.db.WithContext(ctx).Where("recipient_id = ?", recipientID).Order("created_at DESC").Find(&shares).Error
	return shares, err
}

func (r *ShareRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&models.SharedPassword{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"view_count":   gorm.Expr("view_count + ?", 1),
			"last_accessed": time.Now(),
		}).Error
}

func (r *ShareRepository) Revoke(ctx context.Context, token string, ownerID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("share_token = ? AND owner_id = ?", token, ownerID).
		Model(&models.SharedPassword{}).
		Update("revoked", true).
		Error
}
