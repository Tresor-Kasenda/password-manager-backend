package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

type VaultHandler struct {
	vaultRepo     *repository.VaultRepository
	cryptoService *services.CryptoService
}

func NewVaultHandler(
	vaultRepo *repository.VaultRepository,
	cryptoService *services.CryptoService,
) *VaultHandler {
	return &VaultHandler{
		vaultRepo:     vaultRepo,
		cryptoService: cryptoService,
	}
}

func (h *VaultHandler) CreateVault(c *gin.Context) {
	userID := c.GetString("user_id")

	var req models.CreateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	data := models.DecryptedVaultData{
		Password: req.Password,
		Notes:    req.Notes,
	}

	dataJSON, err := json.Marshal(data)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to prepare data"})
		return
	}

	ciphertext, salt, nonce, err := h.cryptoService.EncryptData(
		string(dataJSON),
		req.MasterPassword,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	vault := &models.Vault{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		Title:          req.Title,
		Website:        req.Website,
		Username:       req.Username,
		EncryptedData:  ciphertext,
		EncryptionSalt: salt,
		Nonce:          nonce,
		Folder:         req.Folder,
		Favorite:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.vaultRepo.Create(c.Request.Context(), vault); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create vault entry"})
		return
	}

	c.JSON(http.StatusOK, h.toVaultResponse(vault))
}

func (h *VaultHandler) GetVaults(c *gin.Context) {
	userID := c.GetString("user_id")

	vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
		return
	}

	response := make([]models.VaultResponse, len(vaults))
	for i, vault := range vaults {
		response[i] = h.toVaultResponse(&vault)
	}

	c.JSON(http.StatusOK, response)
}

func (h *VaultHandler) GetVault(c *gin.Context) {
	userID := c.GetString("user_id")
	vaultID := c.Param("id")
	masterPassword := c.Query("master_password")

	if masterPassword == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Master password required"})
		return
	}

	vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vault"})
		return
	}
	if vault == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault entry not found"})
		return
	}

	if vault.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	plaintext, err := h.cryptoService.DecryptData(
		vault.EncryptedData,
		masterPassword,
		vault.EncryptionSalt,
		vault.Nonce,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
		return
	}

	var data models.DecryptedVaultData
	if err := json.Unmarshal([]byte(plaintext), &data); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrypt data"})
		return
	}

	response := h.toVaultResponse(vault)

	c.JSON(http.StatusOK, gin.H{
		"id":         response.ID,
		"title":      response.Title,
		"website":    response.Website,
		"username":   response.Username,
		"password":   data.Password,
		"notes":      data.Notes,
		"folder":     response.Folder,
		"favorite":   response.Favorite,
		"created_at": response.CreatedAt,
		"updated_at": response.UpdatedAt,
	})
}

func (h *VaultHandler) UpdateVault(c *gin.Context) {
	userID := c.GetString("user_id")
	vaultID := c.Param("id")

	var req models.CreateVaultRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vault"})
		return
	}
	if vault == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault entry not found"})
		return
	}

	if vault.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	data := models.DecryptedVaultData{
		Password: req.Password,
		Notes:    req.Notes,
	}

	dataJSON, _ := json.Marshal(data)

	ciphertext, salt, nonce, err := h.cryptoService.EncryptData(
		string(dataJSON),
		req.MasterPassword,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
		return
	}

	vault.Title = req.Title
	vault.Website = req.Website
	vault.Username = req.Username
	vault.EncryptedData = ciphertext
	vault.EncryptionSalt = salt
	vault.Nonce = nonce
	vault.Folder = req.Folder
	vault.UpdatedAt = time.Now()

	if err := h.vaultRepo.Update(c.Request.Context(), vault); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update vault"})
		return
	}

	c.JSON(http.StatusOK, h.toVaultResponse(vault))
}

func (h *VaultHandler) DeleteVault(c *gin.Context) {
	userID := c.GetString("user_id")
	vaultID := c.Param("id")

	vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vault"})
		return
	}
	if vault == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Vault entry not found"})
		return
	}

	if vault.UserID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := h.vaultRepo.Delete(c.Request.Context(), uuid.MustParse(vaultID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vault"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Vault entry deleted successfully"})
}

func (h *VaultHandler) GeneratePassword(c *gin.Context) {
	length := 20
	if l := c.Query("length"); l != "" {
		if _, err := fmt.Sscanf(l, "%d", &length); err == nil {
			if length < 8 {
				length = 8
			} else if length > 128 {
				length = 128
			}
		}
	}

	useSpecial := c.DefaultQuery("use_special", "true") == "true"

	password, err := h.cryptoService.GeneratePassword(length, useSpecial)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"password": password})
}

func (h *VaultHandler) toVaultResponse(vault *models.Vault) models.VaultResponse {
	return models.VaultResponse{
		ID:        vault.ID,
		Title:     vault.Title,
		Website:   vault.Website,
		Username:  vault.Username,
		Folder:    vault.Folder,
		Favorite:  vault.Favorite,
		CreatedAt: vault.CreatedAt,
		UpdatedAt: vault.UpdatedAt,
	}
}
