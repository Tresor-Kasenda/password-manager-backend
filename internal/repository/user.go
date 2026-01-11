package repository

import (
	"context"
	"errors"
	"log"

	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user data persistence with GORM
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user in the database
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	return &user, err
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		log.Printf("GetByEmail query failed for email=%s: %v", email, err)
		return nil, err
	}

	return &user, nil
}

// Update updates an existing user in the database
// GORM automatically updates the UpdatedAt timestamp via the autoUpdateTime tag
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, id).Error
}
