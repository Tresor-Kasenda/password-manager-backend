package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

type HealthHandler struct {
	vaultRepo         *repository.VaultRepository
	cryptoService     *services.CryptoService
	passwordHealthSvc *services.PasswordHealthService
	breachService     *services.BreachService
}

func NewHealthHandler(
	vaultRepo *repository.VaultRepository,
	cryptoService *services.CryptoService,
	passwordHealthSvc *services.PasswordHealthService,
	breachService *services.BreachService,
) *HealthHandler {
	return &HealthHandler{
		vaultRepo:         vaultRepo,
		cryptoService:     cryptoService,
		passwordHealthSvc: passwordHealthSvc,
		breachService:     breachService,
	}
}

func (h *HealthHandler) GetHealthReport(c *gin.Context) {
	userID := c.GetString("user_id")

	vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
		return
	}

	report := h.passwordHealthSvc.GenerateHealthReport(vaults)

	c.JSON(http.StatusOK, report)
}

func (h *HealthHandler) ScanAllPasswords(c *gin.Context) {
	userID := c.GetString("user_id")

	vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
		return
	}

	var vulnerablePasswords []map[string]interface{}

	for _, vault := range vaults {
		// In production, decrypt the password properly
		password := "dummy_password" // Replace with actual decryption

		// Check breach
		breached, count, _ := h.breachService.CheckBreach(password)

		// Calculate strength
		strength := h.passwordHealthSvc.CalculateStrength(password, vault.UpdatedAt)

		if breached || strength["score"].(int) < 60 {
			vulnerablePasswords = append(vulnerablePasswords, map[string]interface{}{
				"vault_id": vault.ID,
				"title":    vault.Title,
				"website":  vault.Website,
				"breach_status": map[string]interface{}{
					"breached": breached,
					"count":    count,
					"message":  getBreachMessage(breached, count),
				},
				"strength": strength,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_scanned":        len(vaults),
		"issues_found":         len(vulnerablePasswords),
		"vulnerable_passwords": vulnerablePasswords,
	})
}

func (h *HealthHandler) AnalyzePassword(c *gin.Context) {
	var req struct {
		Password    string     `json:"password" binding:"required"`
		LastChanged *time.Time `json:"last_changed"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	lastChanged := time.Now()
	if req.LastChanged != nil {
		lastChanged = *req.LastChanged
	}

	strength := h.passwordHealthSvc.CalculateStrength(req.Password, lastChanged)

	c.JSON(http.StatusOK, strength)
}

func (h *HealthHandler) CheckPasswordBreach(c *gin.Context) {
	var req struct {
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	breached, count, err := h.breachService.CheckBreach(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check breach"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"breached": breached,
		"count":    count,
		"message":  getBreachMessage(breached, count),
	})
}

func getBreachMessage(breached bool, count int) string {
	if breached {
		return fmt.Sprintf("This password has appeared %d times in data breaches!", count)
	}
	return "Password has not been found in known breaches"
}
