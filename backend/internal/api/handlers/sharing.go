package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

type SharingHandler struct {
	shareRepo     *repository.ShareRepository
	vaultRepo     *repository.VaultRepository
	userRepo      *repository.UserRepository
	cryptoService *services.CryptoService
	emailService  *services.EmailService
}

func NewSharingHandler(
	shareRepo *repository.ShareRepository,
	vaultRepo *repository.VaultRepository,
	userRepo *repository.UserRepository,
	cryptoService *services.CryptoService,
	emailService *services.EmailService,
) *SharingHandler {
	return &SharingHandler{
		shareRepo:     shareRepo,
		vaultRepo:     vaultRepo,
		userRepo:      userRepo,
		cryptoService: cryptoService,
		emailService:  emailService,
	}
}

func (h *SharingHandler) SharePassword(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.SharePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(req.VaultID))
	if err != nil || vault.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	recipient, err := h.userRepo.GetByEmail(c.Request.Context(), req.RecipientEmail)
	if err != nil || recipient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Recipient not found"})
		return
	}

	shareToken := h.generateShareToken()

	var expiresAt *time.Time
	if req.ExpiresInHours != nil && *req.ExpiresInHours > 0 {
		exp := time.Now().Add(time.Hour * time.Duration(*req.ExpiresInHours))
		expiresAt = &exp
	}

	var sharePasswordHash *string
	if req.RequirePassword && req.SharePassword != nil {
		hash, _, err := h.cryptoService.HashPassword(*req.SharePassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash share password"})
			return
		}
		sharePasswordHash = &hash
	}

	share := &models.SharedPassword{
		ID:                uuid.New(),
		VaultID:           vault.ID,
		OwnerID:           uuid.MustParse(userID),
		RecipientID:       &recipient.ID,
		RecipientEmail:    req.RecipientEmail,
		EncryptedData:     vault.EncryptedData,
		ShareToken:        shareToken,
		ExpiresAt:         expiresAt,
		MaxViews:          req.MaxViews,
		ViewCount:         0,
		RequirePassword:   req.RequirePassword,
		SharePasswordHash: sharePasswordHash,
		CanView:           true,
		CanCopy:           true,
		CanEdit:           req.CanEdit,
		Revoked:           false,
		CreatedAt:         time.Now(),
	}

	if err := h.shareRepo.Create(c.Request.Context(), share); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create share"})
		return
	}

	shareURL := fmt.Sprintf("https://yourdomain.com/shared/%s", shareToken)
	go h.emailService.SendShareNotification(req.RecipientEmail, vault.Title, shareURL)

	c.JSON(http.StatusOK, models.SharePasswordResponse{
		ShareToken: shareToken,
		ShareURL:   shareURL,
		ExpiresAt:  expiresAt,
	})
}

func (h *SharingHandler) GetSharedPassword(c *gin.Context) {
	userID := c.GetString("user_id")
	shareToken := c.Param("token")
	sharePassword := c.Query("share_password")

	share, err := h.shareRepo.GetByToken(c.Request.Context(), shareToken)
	if err != nil || share == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Share not found"})
		return
	}

	if share.Revoked {
		c.JSON(http.StatusGone, gin.H{"error": "Share has been revoked"})
		return
	}

	if share.ExpiresAt != nil && time.Now().After(*share.ExpiresAt) {
		c.JSON(http.StatusGone, gin.H{"error": "Share has expired"})
		return
	}

	if share.MaxViews != nil && share.ViewCount >= *share.MaxViews {
		c.JSON(http.StatusGone, gin.H{"error": "Share view limit reached"})
		return
	}

	if share.RecipientID != nil && share.RecipientID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if share.RequirePassword {
		if sharePassword == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Share password required"})
			return
		}
		if !h.cryptoService.VerifyPassword(sharePassword, *share.SharePasswordHash, "") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid share password"})
			return
		}
	}

	vault, err := h.vaultRepo.GetByID(c.Request.Context(), share.VaultID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve password"})
		return
	}

	err = h.shareRepo.IncrementViewCount(c.Request.Context(), share.ID)
	if err != nil {
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"title":    vault.Title,
			"website":  vault.Website,
			"username": vault.Username,
			"password": "***",
			"notes":    nil,
		},
		"permissions": gin.H{
			"can_view": share.CanView,
			"can_copy": share.CanCopy,
			"can_edit": share.CanEdit,
		},
		"metadata": gin.H{
			"owner":           share.OwnerID,
			"expires_at":      share.ExpiresAt,
			"views_remaining": calculateViewsRemaining(share),
		},
	})
}

func (h *SharingHandler) RevokeShare(c *gin.Context) {
	userID := c.GetString("user_id")
	shareToken := c.Param("token")

	if err := h.shareRepo.Revoke(c.Request.Context(), shareToken, uuid.MustParse(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke share"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Share revoked successfully"})
}

func (h *SharingHandler) ListShares(c *gin.Context) {
	userID := c.GetString("user_id")
	userUUID := uuid.MustParse(userID)

	sent, err := h.shareRepo.GetSentShares(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sent shares"})
		return
	}

	received, err := h.shareRepo.GetReceivedShares(c.Request.Context(), userUUID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch received shares"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"sent":     sent,
		"received": received,
	})
}

func (h *SharingHandler) generateShareToken() string {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return ""
	}
	return base64.URLEncoding.EncodeToString(b)
}

func calculateViewsRemaining(share *models.SharedPassword) *int {
	if share.MaxViews == nil {
		return nil
	}
	remaining := *share.MaxViews - share.ViewCount
	return &remaining
}
