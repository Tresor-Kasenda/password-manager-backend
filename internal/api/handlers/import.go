package handlers

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/models"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

type sessionEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

type ImportHandler struct {
	vaultRepo     *repository.VaultRepository
	importService *services.ImportService
	cryptoService *services.CryptoService
	// in-memory session store to replace Redis for now
	sessions   map[string]sessionEntry
	sessionsMu sync.RWMutex
}

func NewImportHandler(
	vaultRepo *repository.VaultRepository,
	importService *services.ImportService,
	cryptoService *services.CryptoService,
) *ImportHandler {
	return &ImportHandler{
		vaultRepo:     vaultRepo,
		importService: importService,
		cryptoService: cryptoService,
		sessions:      make(map[string]sessionEntry),
	}
}

type UploadRequest struct {
	Content  string `json:"content" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	Source   string `json:"source" binding:"required"`
}

type ImportSessionResponse struct {
	SessionID      string               `json:"session_id"`
	Source         string               `json:"source"`
	TotalEntries   int                  `json:"total_entries"`
	ValidEntries   int                  `json:"valid_entries"`
	InvalidEntries int                  `json:"invalid_entries"`
	Warnings       int                  `json:"warnings"`
	Preview        []models.ImportEntry `json:"preview"`
}

func (h *ImportHandler) UploadFile(c *gin.Context) {
	userID := c.GetString("user_id")

	var req UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entries, err := h.importService.ParseImportFile(req.Content, req.Source)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse import file: " + err.Error()})
		return
	}

	var validEntries []models.ImportEntry
	var invalidEntries []models.ImportEntry
	var warnings []string

	for _, entry := range entries {
		hasTitle := entry.Title != ""
		hasWebsite := entry.Website != nil && *entry.Website != ""

		if !hasTitle && !hasWebsite {
			entry.ValidationIssues = append(entry.ValidationIssues, "Missing title and website")
			invalidEntries = append(invalidEntries, entry)
			continue
		}

		if entry.Password == "" {
			entry.ValidationIssues = append(entry.ValidationIssues, "Missing password")
			invalidEntries = append(invalidEntries, entry)
			continue
		}

		strength := h.cryptoService.CalculatePasswordStrength(entry.Password)
		if strength["score"].(int) < 40 {
			warnings = append(warnings, entry.Title+" has a weak password")
		}

		validEntries = append(validEntries, entry)
	}

	sessionID := uuid.New().String()

	sessionData := map[string]interface{}{
		"user_id":         userID,
		"source":          req.Source,
		"valid_entries":   validEntries,
		"invalid_entries": invalidEntries,
		"warnings":        warnings,
	}

	sessionJSON, _ := json.Marshal(sessionData)

	// store session in-memory with TTL
	h.sessionsMu.Lock()
	h.sessions[sessionID] = sessionEntry{
		Data:      sessionJSON,
		ExpiresAt: time.Now().Add(time.Hour),
	}
	h.sessionsMu.Unlock()

	preview := validEntries
	if len(preview) > 10 {
		preview = preview[:10]
	}

	c.JSON(http.StatusOK, ImportSessionResponse{
		SessionID:      sessionID,
		Source:         req.Source,
		TotalEntries:   len(entries),
		ValidEntries:   len(validEntries),
		InvalidEntries: len(invalidEntries),
		Warnings:       len(warnings),
		Preview:        preview,
	})
}

func (h *ImportHandler) ConfirmImport(c *gin.Context) {
	userID := c.GetString("user_id")
	sessionID := c.Param("session_id")

	var req struct {
		MasterPassword string `json:"master_password" binding:"required"`
		MergeStrategy  string `json:"merge_strategy" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get session from in-memory store
	h.sessionsMu.RLock()
	sessionEntry, ok := h.sessions[sessionID]
	h.sessionsMu.RUnlock()

	if !ok || time.Now().After(sessionEntry.ExpiresAt) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Import session not found or expired"})
		return
	}

	var sessionData map[string]interface{}
	if err := json.Unmarshal(sessionEntry.Data, &sessionData); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse session data"})
		return
	}

	if sessionData["user_id"].(string) != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	validEntriesJSON, _ := json.Marshal(sessionData["valid_entries"])
	var validEntries []models.ImportEntry
	json.Unmarshal(validEntriesJSON, &validEntries)

	imported := 0
	skipped := 0
	var errors []map[string]string

	for _, entry := range validEntries {
		existing, _ := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))

		exists := false
		for _, vault := range existing {
			if vault.Website != nil && entry.Website != nil &&
				*vault.Website == *entry.Website &&
				vault.Username != nil && entry.Username != nil &&
				*vault.Username == *entry.Username {
				exists = true
				break
			}
		}

		if exists {
			switch req.MergeStrategy {
			case "skip":
				skipped++
				continue
			case "overwrite":
				skipped++
				continue
			case "create_new":
				// Continue to create
			}
		}

		dataJSON, _ := json.Marshal(map[string]interface{}{
			"password": entry.Password,
			"notes":    entry.Notes,
		})

		ciphertext, salt, nonce, err := h.cryptoService.EncryptData(
			string(dataJSON),
			req.MasterPassword,
		)

		if err != nil {
			errors = append(errors, map[string]string{
				"title": entry.Title,
				"error": "Encryption failed",
			})
			continue
		}

		vault := &models.Vault{
			ID:             uuid.New(),
			UserID:         uuid.MustParse(userID),
			Title:          entry.Title,
			Website:        entry.Website,
			Username:       entry.Username,
			EncryptedData:  ciphertext,
			EncryptionSalt: salt,
			Nonce:          nonce,
			Folder:         entry.Folder,
			Favorite:       entry.Favorite,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}

		if err := h.vaultRepo.Create(c.Request.Context(), vault); err != nil {
			errors = append(errors, map[string]string{
				"title": entry.Title,
				"error": "Failed to create entry",
			})
			continue
		}

		imported++
	}

	// delete session
	h.sessionsMu.Lock()
	delete(h.sessions, sessionID)
	h.sessionsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{
		"imported":      imported,
		"skipped":       skipped,
		"errors":        len(errors),
		"error_details": errors,
	})
}

func (h *ImportHandler) GetSupportedFormats(c *gin.Context) {
	formats := []map[string]interface{}{
		{
			"id":           "1password",
			"name":         "1Password",
			"file_types":   []string{".csv"},
			"instructions": "Export from 1Password: File → Export → CSV",
		},
		{
			"id":           "lastpass",
			"name":         "LastPass",
			"file_types":   []string{".csv"},
			"instructions": "Export from LastPass: Account Options → Advanced → Export",
		},
		{
			"id":           "bitwarden",
			"name":         "Bitwarden",
			"file_types":   []string{".json"},
			"instructions": "Export from Bitwarden: Tools → Export Vault → JSON",
		},
		{
			"id":           "chrome",
			"name":         "Chrome Browser",
			"file_types":   []string{".csv"},
			"instructions": "Export from Chrome: Settings → Passwords → Export passwords",
		},
		{
			"id":           "keepass",
			"name":         "KeePass",
			"file_types":   []string{".xml"},
			"instructions": "Export from KeePass: File → Export → KeePass XML (2.x)",
		},
	}

	c.JSON(http.StatusOK, gin.H{"formats": formats})
}
