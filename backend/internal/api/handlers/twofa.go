package handlers

import (
	"encoding/base64"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
	"github.com/xlzd/gotp"
)

type TwoFAHandler struct {
	userRepo      *repository.UserRepository
	cryptoService *services.CryptoService
}

func NewTwoFAHandler(
	userRepo *repository.UserRepository,
	cryptoService *services.CryptoService,
) *TwoFAHandler {
	return &TwoFAHandler{
		userRepo:      userRepo,
		cryptoService: cryptoService,
	}
}

func (h *TwoFAHandler) Enable2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MasterPassword string `json:"master_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !h.cryptoService.VerifyPassword(req.MasterPassword, user.MasterPasswordHash, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
		return
	}

	secret := gotp.RandomSecret(32)

	uri := gotp.NewDefaultTOTP(secret).ProvisioningUri(user.Email, "SecureVault")

	qr, err := qrcode.Encode(uri, qrcode.Medium, 256)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate QR code"})
		return
	}

	qrBase64 := "data:image/png;base64," + base64.StdEncoding.EncodeToString(qr)

	backupCodes := h.generateBackupCodes(10)

	secretStr := secret
	user.TwoFactorSecret = &secretStr
	user.BackupCodes = backupCodes
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"secret":       secret,
		"qr_code":      qrBase64,
		"backup_codes": backupCodes,
	})
}

func (h *TwoFAHandler) VerifyAndEnable(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.TwoFactorSecret == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA setup not initiated"})
		return
	}

	totp := gotp.NewDefaultTOTP(*user.TwoFactorSecret)
	if !totp.Verify(req.Token, int64(int(time.Now().Unix()))) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	user.TwoFactorEnabled = true
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to enable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA enabled successfully"})
}

func (h *TwoFAHandler) Verify2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		Token string `json:"token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !user.TwoFactorEnabled || user.TwoFactorSecret == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "2FA not enabled"})
		return
	}

	totp := gotp.NewDefaultTOTP(*user.TwoFactorSecret)
	if totp.Verify(req.Token, int64(int(time.Now().Unix()))) {
		c.JSON(http.StatusOK, gin.H{"valid": true})
		return
	}

	for i, code := range user.BackupCodes {
		if code == req.Token {
			user.BackupCodes = append(user.BackupCodes[:i], user.BackupCodes[i+1:]...)
			err := h.userRepo.Update(c.Request.Context(), user)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update backup codes"})
				return
			}

			c.JSON(http.StatusOK, gin.H{"valid": true, "backup_code_used": true})
			return
		}
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
}

func (h *TwoFAHandler) Disable2FA(c *gin.Context) {
	userID := c.GetString("user_id")

	var req struct {
		MasterPassword string `json:"master_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if !h.cryptoService.VerifyPassword(req.MasterPassword, user.MasterPasswordHash, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
		return
	}

	user.TwoFactorEnabled = false
	user.TwoFactorSecret = nil
	user.BackupCodes = []string{}
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disable 2FA"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "2FA disabled successfully"})
}

func (h *TwoFAHandler) generateBackupCodes(count int) []string {
	codes := make([]string, count)

	for i := 0; i < count; i++ {
		code := ""
		for j := 0; j < 8; j++ {
			if j == 4 {
				code += "-"
			}
			chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
			code += string(chars[rand.Intn(len(chars))])
		}
		codes[i] = code
	}

	return codes
}
