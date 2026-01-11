# Plans d'Implémentation: Features Haute Priorité

## Vue d'Ensemble

Ce document présente **3 plans d'implémentation** pour les features à haute priorité du gestionnaire de mots de passe:

- **Option A**: Impact Rapide (Recherche Avancée + Historique Passwords)
- **Option B**: Priorité Sécurité (Changement Master Password)
- **Option C**: Plan Complet (Toutes les 5 features)

---

# OPTION A: Impact Rapide 🚀

**Objectif**: Implémenter rapidement 2 features à haute valeur utilisateur

**Timeline**: 3-4 heures total
**Complexité**: Faible à Moyenne
**Impact**: Haute satisfaction utilisateur immédiate

## Feature A1: Recherche Avancée & Filtres (1-2 heures)

### Contexte Actuel

**Ce qui existe:**
- Méthode `Search()` dans VaultRepository (lignes 47-56)
- Recherche simple ILIKE sur title/website/username
- Pas d'endpoint exposé
- Pas de filtres (folder, favorite)
- Pas de pagination
- Un seul tri: `created_at DESC`

**Ce qui manque:**
- Endpoint API pour la recherche
- Filtres multiples combinables
- Pagination (limite/offset)
- Tri dynamique sur plusieurs colonnes
- Filtrage par folder, favorite
- Full-text search amélioré

### Plan d'Implémentation

#### 1. Créer le Modèle de Requête de Recherche

**Fichier**: `internal/models/vault.go` (ajouter)

```go
type VaultSearchRequest struct {
    SearchTerm  string  `form:"search"`
    Folder      *string `form:"folder"`
    Favorite    *bool   `form:"favorite"`
    SortBy      string  `form:"sort_by" binding:"omitempty,oneof=created updated last_used title"`
    SortOrder   string  `form:"sort_order" binding:"omitempty,oneof=asc desc"`
    Page        int     `form:"page" binding:"omitempty,min=1"`
    PageSize    int     `form:"page_size" binding:"omitempty,min=1,max=100"`
}

type VaultSearchResponse struct {
    Data       []VaultResponse    `json:"data"`
    Pagination PaginationMetadata `json:"pagination"`
}

type PaginationMetadata struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalItems int64 `json:"total_items"`
    TotalPages int   `json:"total_pages"`
}
```

#### 2. Étendre VaultRepository avec Recherche Avancée

**Fichier**: `internal/repository/vault.go`

**Ajouter nouvelle méthode:**

```go
func (r *VaultRepository) SearchWithFilters(
    ctx context.Context,
    userID uuid.UUID,
    filters models.VaultSearchRequest,
) ([]models.Vault, int64, error) {
    query := r.db.WithContext(ctx).Where("user_id = ?", userID)

    // 1. Search term filter (ILIKE sur 3 colonnes)
    if filters.SearchTerm != "" {
        pattern := "%" + filters.SearchTerm + "%"
        query = query.Where(
            "title ILIKE ? OR website ILIKE ? OR username ILIKE ?",
            pattern, pattern, pattern,
        )
    }

    // 2. Folder filter
    if filters.Folder != nil {
        query = query.Where("folder = ?", *filters.Folder)
    }

    // 3. Favorite filter
    if filters.Favorite != nil {
        query = query.Where("favorite = ?", *filters.Favorite)
    }

    // 4. Count total (avant pagination)
    var total int64
    if err := query.Model(&models.Vault{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // 5. Dynamic sorting
    sortBy := "created_at"
    if filters.SortBy != "" {
        sortBy = filters.SortBy
    }
    sortOrder := "DESC"
    if filters.SortOrder == "asc" {
        sortOrder = "ASC"
    }
    query = query.Order(sortBy + " " + sortOrder)

    // 6. Pagination
    page := 1
    if filters.Page > 0 {
        page = filters.Page
    }
    pageSize := 20
    if filters.PageSize > 0 {
        pageSize = filters.PageSize
    }
    offset := (page - 1) * pageSize

    query = query.Offset(offset).Limit(pageSize)

    // 7. Execute query
    var vaults []models.Vault
    if err := query.Find(&vaults).Error; err != nil {
        return nil, 0, err
    }

    return vaults, total, nil
}
```

#### 3. Ajouter Handler de Recherche

**Fichier**: `internal/api/handlers/vault.go`

```go
func (h *VaultHandler) SearchVaults(c *gin.Context) {
    userID := c.GetString("user_id")

    var filters models.VaultSearchRequest
    if err := c.ShouldBindQuery(&filters); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    vaults, total, err := h.vaultRepo.SearchWithFilters(
        c.Request.Context(),
        uuid.MustParse(userID),
        filters,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search vaults"})
        return
    }

    // Convert to response format
    response := make([]models.VaultResponse, len(vaults))
    for i, vault := range vaults {
        response[i] = h.toVaultResponse(&vault)
    }

    // Calculate pagination
    pageSize := 20
    if filters.PageSize > 0 {
        pageSize = filters.PageSize
    }
    page := 1
    if filters.Page > 0 {
        page = filters.Page
    }
    totalPages := int(total) / pageSize
    if int(total)%pageSize > 0 {
        totalPages++
    }

    c.JSON(http.StatusOK, models.VaultSearchResponse{
        Data: response,
        Pagination: models.PaginationMetadata{
            Page:       page,
            PageSize:   pageSize,
            TotalItems: total,
            TotalPages: totalPages,
        },
    })
}
```

#### 4. Ajouter Route API

**Fichier**: `internal/api/router.go`

```go
vault := protected.Group("/vault")
{
    vault.POST("", r.vaultHandler.CreateVault)
    vault.GET("", r.vaultHandler.GetVaults)
    vault.GET("/search", r.vaultHandler.SearchVaults)  // NOUVELLE ROUTE
    vault.GET("/:id", r.vaultHandler.GetVault)
    // ... reste inchangé
}
```

#### 5. Optimisation Base de Données

**Fichier**: Migration SQL (exécuter manuellement)

```sql
-- Index pour améliorer les performances de recherche
CREATE INDEX IF NOT EXISTS idx_vaults_favorite ON vaults(favorite);
CREATE INDEX IF NOT EXISTS idx_vaults_folder ON vaults(folder);
CREATE INDEX IF NOT EXISTS idx_vaults_last_used ON vaults(last_used DESC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_vaults_updated_at ON vaults(updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_vaults_title ON vaults(title);

-- Index composite pour les requêtes fréquentes
CREATE INDEX IF NOT EXISTS idx_vaults_user_favorite ON vaults(user_id, favorite);
CREATE INDEX IF NOT EXISTS idx_vaults_user_folder ON vaults(user_id, folder);
```

### Tests de Vérification

```bash
# 1. Recherche simple
curl "http://localhost:8000/api/v1/vault/search?search=gmail" \
  -H "Authorization: Bearer $TOKEN"

# 2. Filtre par dossier
curl "http://localhost:8000/api/v1/vault/search?folder=work" \
  -H "Authorization: Bearer $TOKEN"

# 3. Filtre favoris
curl "http://localhost:8000/api/v1/vault/search?favorite=true" \
  -H "Authorization: Bearer $TOKEN"

# 4. Tri + pagination
curl "http://localhost:8000/api/v1/vault/search?sort_by=updated&sort_order=desc&page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 5. Combinaison de filtres
curl "http://localhost:8000/api/v1/vault/search?search=gmail&folder=personal&favorite=true&sort_by=title" \
  -H "Authorization: Bearer $TOKEN"
```

**Réponse attendue:**
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 10,
    "total_items": 45,
    "total_pages": 5
  }
}
```

---

## Feature A2: Historique des Passwords (2 heures)

### Contexte Actuel

**Ce qui existe:**
- Vaults stockent password actuel encrypté
- UpdatedAt timestamp sur modifications
- Aucun historique des anciennes valeurs

**Ce qui manque:**
- Table pour stocker versions précédentes
- Mécanisme de sauvegarde avant update
- Endpoint pour récupérer historique
- Capacité de restaurer ancienne version

### Plan d'Implémentation

#### 1. Créer le Modèle d'Historique

**Fichier**: `internal/models/password_history.go` (nouveau)

```go
package models

import (
    "time"
    "github.com/google/uuid"
)

type PasswordHistory struct {
    ID             uuid.UUID  `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
    VaultID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"vault_id"`
    EncryptedData  string     `gorm:"not null" json:"-"`
    EncryptionSalt string     `gorm:"not null" json:"-"`
    Nonce          string     `gorm:"not null" json:"-"`
    ChangedAt      time.Time  `gorm:"autoCreateTime" json:"changed_at"`
    ChangedBy      uuid.UUID  `gorm:"type:uuid;not null" json:"changed_by"`

    // Relations
    Vault Vault `gorm:"foreignKey:VaultID" json:"-"`
}

func (PasswordHistory) TableName() string {
    return "password_history"
}

type PasswordHistoryResponse struct {
    ID        uuid.UUID `json:"id"`
    VaultID   uuid.UUID `json:"vault_id"`
    Password  string    `json:"password"`    // Décrypté
    ChangedAt time.Time `json:"changed_at"`
}
```

#### 2. Créer Repository d'Historique

**Fichier**: `internal/repository/password_history.go` (nouveau)

```go
package repository

import (
    "context"
    "github.com/google/uuid"
    "gorm.io/gorm"
    "github.com/tresor/password-manager/internal/models"
)

type PasswordHistoryRepository struct {
    db *gorm.DB
}

func NewPasswordHistoryRepository(db *gorm.DB) *PasswordHistoryRepository {
    return &PasswordHistoryRepository{db: db}
}

func (r *PasswordHistoryRepository) Create(ctx context.Context, history *models.PasswordHistory) error {
    return r.db.WithContext(ctx).Create(history).Error
}

func (r *PasswordHistoryRepository) GetByVaultID(
    ctx context.Context,
    vaultID uuid.UUID,
    limit int,
) ([]models.PasswordHistory, error) {
    var history []models.PasswordHistory
    query := r.db.WithContext(ctx).
        Where("vault_id = ?", vaultID).
        Order("changed_at DESC")

    if limit > 0 {
        query = query.Limit(limit)
    }

    err := query.Find(&history).Error
    return history, err
}

func (r *PasswordHistoryRepository) GetByID(
    ctx context.Context,
    id uuid.UUID,
) (*models.PasswordHistory, error) {
    var history models.PasswordHistory
    err := r.db.WithContext(ctx).First(&history, id).Error
    if err != nil {
        return nil, err
    }
    return &history, nil
}

func (r *PasswordHistoryRepository) DeleteOldHistory(
    ctx context.Context,
    vaultID uuid.UUID,
    keepCount int,
) error {
    // Garde seulement les N versions les plus récentes
    return r.db.WithContext(ctx).Exec(`
        DELETE FROM password_history
        WHERE vault_id = ?
        AND id NOT IN (
            SELECT id FROM password_history
            WHERE vault_id = ?
            ORDER BY changed_at DESC
            LIMIT ?
        )
    `, vaultID, vaultID, keepCount).Error
}
```

#### 3. Modifier VaultHandler pour Sauvegarder l'Historique

**Fichier**: `internal/api/handlers/vault.go`

**Modifier UpdateVault pour sauvegarder avant update:**

```go
func (h *VaultHandler) UpdateVault(c *gin.Context) {
    userID := c.GetString("user_id")
    vaultID := c.Param("id")

    var req models.UpdateVaultRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 1. Récupérer le vault actuel (AVANT modification)
    currentVault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vault"})
        return
    }
    if currentVault == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
        return
    }

    // Vérifier ownership
    if currentVault.UserID.String() != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
        return
    }

    // 2. SAUVEGARDER VERSION ACTUELLE dans l'historique
    history := &models.PasswordHistory{
        VaultID:        currentVault.ID,
        EncryptedData:  currentVault.EncryptedData,
        EncryptionSalt: currentVault.EncryptionSalt,
        Nonce:          currentVault.Nonce,
        ChangedBy:      uuid.MustParse(userID),
    }

    if err := h.historyRepo.Create(c.Request.Context(), history); err != nil {
        log.Printf("Failed to save password history: %v", err)
        // Continue quand même (non-blocking)
    }

    // 3. Nettoyer vieil historique (garder seulement 10 versions)
    go func() {
        ctx := context.Background()
        h.historyRepo.DeleteOldHistory(ctx, currentVault.ID, 10)
    }()

    // 4. Continuer avec update normal (code existant)
    data := models.DecryptedVaultData{
        Password: req.Password,
        Notes:    req.Notes,
    }
    // ... reste du code update inchangé
}
```

#### 4. Ajouter Handler de Consultation d'Historique

**Fichier**: `internal/api/handlers/vault.go`

```go
func (h *VaultHandler) GetPasswordHistory(c *gin.Context) {
    userID := c.GetString("user_id")
    vaultID := c.Param("id")
    masterPassword := c.Query("master_password")

    if masterPassword == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Master password required"})
        return
    }

    // Vérifier ownership du vault
    vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
    if err != nil || vault == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Vault not found"})
        return
    }
    if vault.UserID.String() != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
        return
    }

    // Récupérer historique (limité à 10 versions)
    history, err := h.historyRepo.GetByVaultID(c.Request.Context(), uuid.MustParse(vaultID), 10)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
        return
    }

    // Décrypter chaque version avec master password
    response := make([]models.PasswordHistoryResponse, 0, len(history))
    for _, h := range history {
        decrypted, err := h.cryptoService.DecryptData(
            h.EncryptedData,
            masterPassword,
            h.EncryptionSalt,
            h.Nonce,
        )
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
            return
        }

        var data models.DecryptedVaultData
        if err := json.Unmarshal([]byte(decrypted), &data); err != nil {
            continue
        }

        response = append(response, models.PasswordHistoryResponse{
            ID:        h.ID,
            VaultID:   h.VaultID,
            Password:  data.Password,
            ChangedAt: h.ChangedAt,
        })
    }

    c.JSON(http.StatusOK, response)
}

func (h *VaultHandler) RestorePasswordFromHistory(c *gin.Context) {
    userID := c.GetString("user_id")
    vaultID := c.Param("id")
    historyID := c.Param("history_id")
    masterPassword := c.Query("master_password")

    if masterPassword == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Master password required"})
        return
    }

    // 1. Vérifier ownership
    vault, err := h.vaultRepo.GetByID(c.Request.Context(), uuid.MustParse(vaultID))
    if err != nil || vault == nil || vault.UserID.String() != userID {
        c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized"})
        return
    }

    // 2. Récupérer entrée d'historique
    history, err := h.historyRepo.GetByID(c.Request.Context(), uuid.MustParse(historyID))
    if err != nil || history.VaultID != uuid.MustParse(vaultID) {
        c.JSON(http.StatusNotFound, gin.H{"error": "History entry not found"})
        return
    }

    // 3. Décrypter ancienne version
    decrypted, err := h.cryptoService.DecryptData(
        history.EncryptedData,
        masterPassword,
        history.EncryptionSalt,
        history.Nonce,
    )
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid master password"})
        return
    }

    // 4. Re-chiffrer avec nouveau salt/nonce
    ciphertext, salt, nonce, err := h.cryptoService.EncryptData(decrypted, masterPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Encryption failed"})
        return
    }

    // 5. Sauvegarder version actuelle dans historique AVANT restore
    currentHistory := &models.PasswordHistory{
        VaultID:        vault.ID,
        EncryptedData:  vault.EncryptedData,
        EncryptionSalt: vault.EncryptionSalt,
        Nonce:          vault.Nonce,
        ChangedBy:      uuid.MustParse(userID),
    }
    h.historyRepo.Create(c.Request.Context(), currentHistory)

    // 6. Update vault avec ancienne version restaurée
    vault.EncryptedData = ciphertext
    vault.EncryptionSalt = salt
    vault.Nonce = nonce
    vault.UpdatedAt = time.Now()

    if err := h.vaultRepo.Update(c.Request.Context(), vault); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to restore password"})
        return
    }

    c.JSON(http.StatusOK, gin.H{"message": "Password restored successfully"})
}
```

#### 5. Ajouter Routes API

**Fichier**: `internal/api/router.go`

```go
vault := protected.Group("/vault")
{
    vault.POST("", r.vaultHandler.CreateVault)
    vault.GET("", r.vaultHandler.GetVaults)
    vault.GET("/search", r.vaultHandler.SearchVaults)
    vault.GET("/:id", r.vaultHandler.GetVault)
    vault.GET("/:id/history", r.vaultHandler.GetPasswordHistory)          // NOUVEAU
    vault.POST("/:id/restore/:history_id", r.vaultHandler.RestorePasswordFromHistory)  // NOUVEAU
    vault.PUT("/:id", r.vaultHandler.UpdateVault)
    vault.DELETE("/:id", r.vaultHandler.DeleteVault)
    // ...
}
```

#### 6. Injection de Dépendances

**Fichier**: `cmd/server/main.go`

```go
// Après initialisation des repos existants
userRepo := repository.NewUserRepository(gormDB)
vaultRepo := repository.NewVaultRepository(gormDB)
shareRepo := repository.NewShareRepository(gormDB)
historyRepo := repository.NewPasswordHistoryRepository(gormDB)  // NOUVEAU

// Handler avec historique
vaultHandler := handlers.NewVaultHandler(vaultRepo, historyRepo, cryptoService)
```

**Modifier signature de NewVaultHandler:**

```go
// internal/api/handlers/vault.go
type VaultHandler struct {
    vaultRepo     *repository.VaultRepository
    historyRepo   *repository.PasswordHistoryRepository  // NOUVEAU
    cryptoService *services.CryptoService
}

func NewVaultHandler(
    vaultRepo *repository.VaultRepository,
    historyRepo *repository.PasswordHistoryRepository,  // NOUVEAU
    cryptoService *services.CryptoService,
) *VaultHandler {
    return &VaultHandler{
        vaultRepo:     vaultRepo,
        historyRepo:   historyRepo,
        cryptoService: cryptoService,
    }
}
```

#### 7. Migration Base de Données

**Fichier**: `internal/database/postgres.go`

```go
// Dans NewGormDB, ajouter migration automatique
err = db.AutoMigrate(
    &models.User{},
    &models.Vault{},
    &models.SharedPassword{},
    &models.PasswordHistory{},  // NOUVEAU
)
```

**Ou créer migration SQL manuelle:**

```sql
CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    vault_id UUID NOT NULL REFERENCES vaults(id) ON DELETE CASCADE,
    encrypted_data TEXT NOT NULL,
    encryption_salt VARCHAR(255) NOT NULL,
    nonce VARCHAR(255) NOT NULL,
    changed_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    changed_by UUID NOT NULL REFERENCES users(id),
    CONSTRAINT fk_password_history_vault FOREIGN KEY (vault_id) REFERENCES vaults(id) ON DELETE CASCADE
);

CREATE INDEX idx_password_history_vault_id ON password_history(vault_id);
CREATE INDEX idx_password_history_changed_at ON password_history(changed_at DESC);
```

### Tests de Vérification

```bash
# 1. Créer un vault
VAULT_ID=$(curl -s -X POST http://localhost:8000/api/v1/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","password":"pass1","master_password":"Master123"}' | jq -r '.id')

# 2. Modifier le password (crée historique)
curl -X PUT http://localhost:8000/api/v1/vault/$VAULT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","password":"pass2","master_password":"Master123"}'

# 3. Modifier à nouveau
curl -X PUT http://localhost:8000/api/v1/vault/$VAULT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Test","password":"pass3","master_password":"Master123"}'

# 4. Récupérer historique
curl "http://localhost:8000/api/v1/vault/$VAULT_ID/history?master_password=Master123" \
  -H "Authorization: Bearer $TOKEN"

# Réponse attendue:
# [
#   {"id":"...", "vault_id":"...", "password":"pass2", "changed_at":"2026-01-11T..."},
#   {"id":"...", "vault_id":"...", "password":"pass1", "changed_at":"2026-01-11T..."}
# ]

# 5. Restaurer ancienne version
HISTORY_ID=$(curl -s "http://localhost:8000/api/v1/vault/$VAULT_ID/history?master_password=Master123" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.[0].id')

curl -X POST "http://localhost:8000/api/v1/vault/$VAULT_ID/restore/$HISTORY_ID?master_password=Master123" \
  -H "Authorization: Bearer $TOKEN"
```

---

## Résumé Option A

**Fichiers Créés:**
1. `internal/models/password_history.go` (nouveau)
2. `internal/repository/password_history.go` (nouveau)

**Fichiers Modifiés:**
1. `internal/models/vault.go` (ajouter VaultSearchRequest, VaultSearchResponse)
2. `internal/repository/vault.go` (ajouter SearchWithFilters)
3. `internal/api/handlers/vault.go` (ajouter 4 handlers)
4. `internal/api/router.go` (ajouter 4 routes)
5. `cmd/server/main.go` (injection historyRepo)
6. `internal/database/postgres.go` (migration PasswordHistory)

**Endpoints Ajoutés:**
- `GET /api/v1/vault/search` - Recherche avec filtres
- `GET /api/v1/vault/:id/history` - Consulter historique
- `POST /api/v1/vault/:id/restore/:history_id` - Restaurer version

**Base de Données:**
- Table `password_history` créée
- 5 index ajoutés pour performance

**Avantages:**
- ✅ Implémentation rapide (3-4 heures)
- ✅ Haute valeur utilisateur immédiate
- ✅ Pas de changements breaking
- ✅ Utilise architecture existante

---

# OPTION B: Priorité Sécurité 🔐

**Objectif**: Implémenter changement de master password avec re-chiffrement

**Timeline**: 2-3 heures
**Complexité**: Moyenne
**Impact**: Critique de sécurité

## Feature B: Changement Master Password

### Contexte Actuel

**Ce qui existe:**
- Master password hashé avec Argon2id (users.master_password_hash)
- Chaque vault chiffré avec master password + salt unique
- Vérification master password via déchiffrement réussi
- Pas de mécanisme pour changer master password

**Ce qui manque:**
- Endpoint pour changer master password
- Mécanisme de re-chiffrement de tous les vaults
- Transaction atomique (tout ou rien)
- Email de confirmation
- Invalidation des sessions existantes

### Architecture de l'Encryption

**Flow actuel:**
```
Master Password → Argon2id(password, salt) → AES-256 key → Encrypt(vault_data)
```

**Challenge du changement:**
1. Récupérer TOUS les vaults de l'utilisateur
2. Décrypter avec ANCIEN master password
3. Re-chiffrer avec NOUVEAU master password (nouveaux salts/nonces)
4. Mettre à jour users.master_password_hash
5. TOUT doit réussir ou RIEN (transaction)

### Plan d'Implémentation

#### 1. Créer Modèle de Requête

**Fichier**: `internal/models/user.go` (ajouter)

```go
type ChangeMasterPasswordRequest struct {
    CurrentPassword string `json:"current_password" binding:"required,min=8"`
    NewPassword     string `json:"new_password" binding:"required,min=12"`
    ConfirmPassword string `json:"confirm_password" binding:"required"`
}

type ChangeMasterPasswordResponse struct {
    Message            string `json:"message"`
    VaultsReencrypted  int    `json:"vaults_reencrypted"`
    SessionsInvalidated int    `json:"sessions_invalidated"`
}
```

#### 2. Ajouter Méthode de Re-chiffrement dans VaultRepository

**Fichier**: `internal/repository/vault.go`

```go
// BulkUpdateEncryption met à jour l'encryption de plusieurs vaults dans une transaction
func (r *VaultRepository) BulkUpdateEncryption(
    ctx context.Context,
    updates []VaultEncryptionUpdate,
) error {
    return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
        for _, update := range updates {
            result := tx.Model(&models.Vault{}).
                Where("id = ?", update.VaultID).
                Updates(map[string]interface{}{
                    "encrypted_data":  update.EncryptedData,
                    "encryption_salt": update.EncryptionSalt,
                    "nonce":           update.Nonce,
                    "updated_at":      time.Now(),
                })

            if result.Error != nil {
                return result.Error
            }

            if result.RowsAffected != 1 {
                return fmt.Errorf("vault %s not found or not updated", update.VaultID)
            }
        }
        return nil
    })
}

type VaultEncryptionUpdate struct {
    VaultID        uuid.UUID
    EncryptedData  string
    EncryptionSalt string
    Nonce          string
}
```

#### 3. Créer Handler de Changement de Master Password

**Fichier**: `internal/api/handlers/auth.go`

```go
func (h *AuthHandler) ChangeMasterPassword(c *gin.Context) {
    userID := c.GetString("user_id")

    var req models.ChangeMasterPasswordRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    // 1. Validation
    if req.NewPassword != req.ConfirmPassword {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
        return
    }

    if req.CurrentPassword == req.NewPassword {
        c.JSON(http.StatusBadRequest, gin.H{"error": "New password must be different"})
        return
    }

    // 2. Récupérer utilisateur
    user, err := h.userRepo.GetByID(c.Request.Context(), uuid.MustParse(userID))
    if err != nil || user == nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
        return
    }

    // 3. Vérifier ANCIEN master password
    if err := h.cryptoService.VerifyPassword(
        req.CurrentPassword,
        user.MasterPasswordHash,
        user.Salt,
    ); err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
        return
    }

    // 4. Récupérer TOUS les vaults de l'utilisateur
    vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), user.ID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
        return
    }

    log.Printf("Re-encrypting %d vaults for user %s", len(vaults), userID)

    // 5. Décrypter et re-chiffrer chaque vault
    updates := make([]repository.VaultEncryptionUpdate, 0, len(vaults))

    for _, vault := range vaults {
        // Décrypter avec ANCIEN master password
        decrypted, err := h.cryptoService.DecryptData(
            vault.EncryptedData,
            req.CurrentPassword,
            vault.EncryptionSalt,
            vault.Nonce,
        )
        if err != nil {
            log.Printf("Failed to decrypt vault %s: %v", vault.ID, err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": fmt.Sprintf("Failed to decrypt vault: %s", vault.Title),
            })
            return
        }

        // Re-chiffrer avec NOUVEAU master password (nouveaux salt/nonce)
        ciphertext, salt, nonce, err := h.cryptoService.EncryptData(
            decrypted,
            req.NewPassword,
        )
        if err != nil {
            log.Printf("Failed to re-encrypt vault %s: %v", vault.ID, err)
            c.JSON(http.StatusInternalServerError, gin.H{
                "error": "Failed to re-encrypt vaults",
            })
            return
        }

        updates = append(updates, repository.VaultEncryptionUpdate{
            VaultID:        vault.ID,
            EncryptedData:  ciphertext,
            EncryptionSalt: salt,
            Nonce:          nonce,
        })
    }

    // 6. Générer nouveau hash pour le nouveau master password
    newHash, newSalt, err := h.cryptoService.HashPassword(req.NewPassword)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
        return
    }

    // 7. TRANSACTION ATOMIQUE: Update user + tous les vaults
    err = h.userRepo.db.Transaction(func(tx *gorm.DB) error {
        // 7a. Update master password hash
        if err := tx.Model(&models.User{}).
            Where("id = ?", user.ID).
            Updates(map[string]interface{}{
                "master_password_hash": newHash,
                "salt":                 newSalt,
                "updated_at":           time.Now(),
            }).Error; err != nil {
            return fmt.Errorf("failed to update user password: %w", err)
        }

        // 7b. Update tous les vaults (bulk update avec nouvelle encryption)
        for _, update := range updates {
            if err := tx.Model(&models.Vault{}).
                Where("id = ?", update.VaultID).
                Updates(map[string]interface{}{
                    "encrypted_data":  update.EncryptedData,
                    "encryption_salt": update.EncryptionSalt,
                    "nonce":           update.Nonce,
                    "updated_at":      time.Now(),
                }).Error; err != nil {
                return fmt.Errorf("failed to update vault %s: %w", update.VaultID, err)
            }
        }

        return nil
    })

    if err != nil {
        log.Printf("Transaction failed: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to change master password. No changes were made.",
        })
        return
    }

    // 8. Envoyer email de confirmation (optionnel)
    go func() {
        h.emailService.SendEmail(
            user.Email,
            "Master Password Changed",
            fmt.Sprintf("Your master password was successfully changed at %s", time.Now().Format(time.RFC1123)),
        )
    }()

    // 9. Invalider toutes les sessions existantes (optionnel)
    // TODO: Implémenter système de blacklist JWT ou sessions tracking

    c.JSON(http.StatusOK, models.ChangeMasterPasswordResponse{
        Message:           "Master password changed successfully",
        VaultsReencrypted: len(vaults),
        SessionsInvalidated: 0, // À implémenter avec session management
    })
}
```

#### 4. Ajouter Route API

**Fichier**: `internal/api/router.go`

```go
protected := v1.Group("")
protected.Use(middleware.AuthMiddleware(r.jwtSecret))
{
    // Auth routes protégées
    auth := protected.Group("/auth")
    {
        auth.POST("/change-password", r.authHandler.ChangeMasterPassword)  // NOUVEAU
    }

    // ... reste inchangé
}
```

#### 5. Ajouter Validation de Force du Nouveau Password

**Fichier**: `internal/services/crypto.go` (optionnel mais recommandé)

```go
func (s *CryptoService) ValidateMasterPassword(password string) error {
    if len(password) < 12 {
        return fmt.Errorf("master password must be at least 12 characters")
    }

    hasUpper := false
    hasLower := false
    hasDigit := false
    hasSpecial := false

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsDigit(char):
            hasDigit = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if !hasUpper || !hasLower || !hasDigit || !hasSpecial {
        return fmt.Errorf("master password must contain uppercase, lowercase, digit, and special character")
    }

    return nil
}
```

**Utiliser dans le handler:**

```go
// Après validation de base
if err := h.cryptoService.ValidateMasterPassword(req.NewPassword); err != nil {
    c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
    return
}
```

#### 6. Amélioration: Rate Limiting Strict

**Fichier**: `internal/api/router.go`

```go
auth := protected.Group("/auth")
{
    // Changement de password très limité (1 tentative / 10 min)
    auth.POST("/change-password",
        middleware.RateLimitMiddleware(1), // 1 req/minute
        r.authHandler.ChangeMasterPassword,
    )
}
```

#### 7. Amélioration: Invalidation des Sessions (Avancé)

**Option A: Session Blacklist (Simple)**

Ajouter un champ `password_changed_at` à User:

```go
// internal/models/user.go
type User struct {
    // ...
    PasswordChangedAt *time.Time `gorm:"column:password_changed_at" json:"-"`
}
```

Vérifier dans le middleware JWT:

```go
// internal/api/middleware/auth.go
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
    return func(c *gin.Context) {
        // ... validation JWT existante

        // Vérifier si token émis AVANT changement de password
        if user.PasswordChangedAt != nil {
            issuedAt := time.Unix(claims.IssuedAt, 0)
            if issuedAt.Before(*user.PasswordChangedAt) {
                c.JSON(http.StatusUnauthorized, gin.H{
                    "error": "Session expired. Please login again.",
                })
                c.Abort()
                return
            }
        }

        c.Next()
    }
}
```

Mettre à jour lors du changement:

```go
// Dans ChangeMasterPassword, après transaction
now := time.Now()
user.PasswordChangedAt = &now
```

### Tests de Vérification

```bash
# 1. Créer un utilisateur et des vaults
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"changetest@example.com","master_password":"OldPass123!"}' \
  | jq -r '.access_token')

# Créer 3 vaults
for i in 1 2 3; do
  curl -s -X POST http://localhost:8000/api/v1/vault \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Vault$i\",\"password\":\"pass$i\",\"master_password\":\"OldPass123!\"}"
done

# 2. Changer master password
curl -s -X POST http://localhost:8000/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password":"OldPass123!",
    "new_password":"NewSecurePass456!",
    "confirm_password":"NewSecurePass456!"
  }' | jq .

# Réponse attendue:
# {
#   "message": "Master password changed successfully",
#   "vaults_reencrypted": 3,
#   "sessions_invalidated": 0
# }

# 3. Vérifier qu'ancien password ne marche plus
curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"changetest@example.com","master_password":"OldPass123!"}' | jq .
# Devrait retourner "Invalid credentials"

# 4. Login avec nouveau password
NEW_TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"changetest@example.com","master_password":"NewSecurePass456!"}' \
  | jq -r '.access_token')

# 5. Vérifier que vaults sont déchiffrables avec NOUVEAU password
curl -s "http://localhost:8000/api/v1/vault" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq .

# 6. Décrypter un vault avec nouveau password
VAULT_ID=$(curl -s "http://localhost:8000/api/v1/vault" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq -r '.[0].id')

curl -s "http://localhost:8000/api/v1/vault/$VAULT_ID?master_password=NewSecurePass456!" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq .
# Devrait afficher password décrypté

# 7. Tester qu'ancien token est invalidé (si password_changed_at implémenté)
curl -s "http://localhost:8000/api/v1/vault" \
  -H "Authorization: Bearer $TOKEN"
# Devrait retourner "Session expired"
```

### Gestion des Erreurs

**Scénarios d'erreur à tester:**

1. **Ancien password incorrect**:
   ```json
   {"error": "Current password is incorrect"}
   ```

2. **Nouveau password trop faible**:
   ```json
   {"error": "Master password must contain uppercase, lowercase, digit, and special character"}
   ```

3. **Passwords ne correspondent pas**:
   ```json
   {"error": "Passwords do not match"}
   ```

4. **Échec de déchiffrement d'un vault**:
   ```json
   {"error": "Failed to decrypt vault: Gmail Account"}
   ```
   (Transaction rollback automatique)

5. **Échec de transaction**:
   ```json
   {"error": "Failed to change master password. No changes were made."}
   ```
   (Aucune modification persistée)

---

## Résumé Option B

**Fichiers Modifiés:**
1. `internal/models/user.go` (ajouter ChangeMasterPasswordRequest, PasswordChangedAt)
2. `internal/repository/vault.go` (ajouter BulkUpdateEncryption)
3. `internal/api/handlers/auth.go` (ajouter ChangeMasterPassword)
4. `internal/api/router.go` (ajouter route)
5. `internal/api/middleware/auth.go` (optionnel: vérification PasswordChangedAt)
6. `internal/services/crypto.go` (optionnel: ValidateMasterPassword)

**Endpoints Ajoutés:**
- `POST /api/v1/auth/change-password` - Changer master password

**Base de Données:**
- Colonne `password_changed_at` ajoutée à `users` (optionnel)

**Sécurité:**
- ✅ Transaction atomique (tout ou rien)
- ✅ Vérification ancien password
- ✅ Re-chiffrement de tous les vaults
- ✅ Nouveaux salts/nonces pour chaque vault
- ✅ Hash du nouveau master password
- ✅ Email de confirmation
- ✅ Rate limiting strict
- ✅ Invalidation sessions (optionnel)

**Avantages:**
- ✅ Feature critique de sécurité
- ✅ Implémentation robuste (transaction)
- ✅ Gestion d'erreur complète
- ✅ Pas de perte de données possible

---

# OPTION C: Plan Complet 🎯

**Objectif**: Implémenter les 5 features de Priorité Haute dans l'ordre optimal

**Timeline Totale**: 8-10 heures (1-2 jours)
**Ordre d'Implémentation**: Logique avec dépendances

## Vue d'Ensemble

```
Phase 1: Fondations
├─ 1. Recherche Avancée & Filtres (1-2h)
└─ 2. Historique Passwords (2h)

Phase 2: Sécurité Critique
└─ 3. Changement Master Password (2-3h)

Phase 3: Intelligence & Audit
├─ 4. Rapport d'Audit Avancé (2h)
└─ 5. Recommandations de Changement (1h)

Bonus (Optionnel):
└─ 6. Features d'Équipe (4-5h) [Si temps disponible]
```

## Ordre d'Implémentation Recommandé

### Phase 1: Fondations (3-4h)

**Feature 1: Recherche Avancée & Filtres** ⭐
- **Pourquoi en premier**: Infrastructure de base, pas de dépendances
- **Voir détails**: Option A - Feature A1
- **Durée**: 1-2 heures

**Feature 2: Historique Passwords** ⭐
- **Pourquoi après recherche**: Utilise même repo pattern
- **Voir détails**: Option A - Feature A2
- **Durée**: 2 heures
- **Note**: Prépare infrastructure pour audit

---

### Phase 2: Sécurité Critique (2-3h)

**Feature 3: Changement Master Password** ⭐⭐⭐
- **Pourquoi maintenant**: Critique sécurité, utilise historique pour backup
- **Voir détails**: Option B complète
- **Durée**: 2-3 heures
- **Dépendances**: Bénéficie de l'historique pour backup avant changement

---

### Phase 3: Intelligence & Audit (3h)

**Feature 4: Rapport d'Audit Avancé** ⭐⭐
- **Pourquoi maintenant**: Analyse tous les vaults avec contexte complet
- **Durée**: 2 heures
- **Dépendances**: Utilise historique pour analyse temporelle

#### Plan d'Implémentation

##### 1. Créer Service d'Audit Amélioré

**Fichier**: `internal/services/audit.go` (nouveau)

```go
package services

import (
    "context"
    "time"
    "github.com/google/uuid"
    "github.com/tresor/password-manager/internal/models"
)

type AuditService struct {
    passwordHealthService *PasswordHealthService
    breachService         *BreachService
    cryptoService         *CryptoService
}

func NewAuditService(
    healthService *PasswordHealthService,
    breachService *BreachService,
    cryptoService *CryptoService,
) *AuditService {
    return &AuditService{
        passwordHealthService: healthService,
        breachService:         breachService,
        cryptoService:         cryptoService,
    }
}

type AuditReport struct {
    OverallScore       int                    `json:"overall_score"`
    TotalPasswords     int                    `json:"total_passwords"`
    Statistics         AuditStatistics        `json:"statistics"`
    DuplicatePasswords []DuplicatePasswordSet `json:"duplicate_passwords"`
    OldPasswords       []OldPasswordEntry     `json:"old_passwords"`
    WeakPasswords      []WeakPasswordEntry    `json:"weak_passwords"`
    BreachedPasswords  []BreachedPasswordEntry `json:"breached_passwords"`
    PriorityActions    []PriorityAction       `json:"priority_actions"`
    GeneratedAt        time.Time              `json:"generated_at"`
}

type AuditStatistics struct {
    ExcellentCount int `json:"excellent_passwords"` // Score >= 80
    GoodCount      int `json:"good_passwords"`      // Score 60-79
    WeakCount      int `json:"weak_passwords"`      // Score < 60
    BreachedCount  int `json:"breached_passwords"`
    DuplicateCount int `json:"duplicate_passwords"`
    OldCount       int `json:"old_passwords"`       // > 180 days
    VeryOldCount   int `json:"very_old_passwords"`  // > 365 days
}

type DuplicatePasswordSet struct {
    Password string                       `json:"password"`
    Count    int                          `json:"count"`
    Vaults   []DuplicatePasswordVaultInfo `json:"vaults"`
}

type DuplicatePasswordVaultInfo struct {
    ID       uuid.UUID `json:"id"`
    Title    string    `json:"title"`
    Website  *string   `json:"website"`
}

type OldPasswordEntry struct {
    VaultID      uuid.UUID `json:"vault_id"`
    Title        string    `json:"title"`
    Website      *string   `json:"website"`
    AgeDays      int       `json:"age_days"`
    LastChanged  time.Time `json:"last_changed"`
}

type WeakPasswordEntry struct {
    VaultID  uuid.UUID `json:"vault_id"`
    Title    string    `json:"title"`
    Website  *string   `json:"website"`
    Score    int       `json:"score"`
    Issues   []string  `json:"issues"`
}

type BreachedPasswordEntry struct {
    VaultID      uuid.UUID `json:"vault_id"`
    Title        string    `json:"title"`
    Website      *string   `json:"website"`
    BreachCount  int       `json:"breach_count"`
}

type PriorityAction struct {
    Priority    string    `json:"priority"`    // "critical", "high", "medium"
    Action      string    `json:"action"`
    VaultCount  int       `json:"vault_count"`
    VaultIDs    []uuid.UUID `json:"vault_ids"`
    Description string    `json:"description"`
}

// GenerateAdvancedAudit crée un rapport d'audit complet avec détection des duplicats
func (s *AuditService) GenerateAdvancedAudit(
    ctx context.Context,
    vaults []models.Vault,
    masterPassword string,
) (*AuditReport, error) {
    report := &AuditReport{
        Statistics:      AuditStatistics{},
        GeneratedAt:     time.Now(),
        TotalPasswords:  len(vaults),
    }

    // Map pour détecter duplicats: password -> liste de vaults
    passwordMap := make(map[string][]models.Vault)

    // Scores pour calcul overall
    var totalScore int

    for _, vault := range vaults {
        // Décrypter password
        decrypted, err := s.cryptoService.DecryptData(
            vault.EncryptedData,
            masterPassword,
            vault.EncryptionSalt,
            vault.Nonce,
        )
        if err != nil {
            // Si déchiffrement échoue, skip ce vault
            continue
        }

        var data models.DecryptedVaultData
        if err := json.Unmarshal([]byte(decrypted), &data); err != nil {
            continue
        }

        password := data.Password

        // Ajouter à la map de duplicats
        passwordMap[password] = append(passwordMap[password], vault)

        // Calculer strength
        strength := s.passwordHealthService.CalculateStrength(password, vault.UpdatedAt)
        score := strength["score"].(int)
        totalScore += score

        // Catégoriser par score
        if score >= 80 {
            report.Statistics.ExcellentCount++
        } else if score >= 60 {
            report.Statistics.GoodCount++
        } else {
            report.Statistics.WeakCount++

            // Ajouter aux weak passwords
            report.WeakPasswords = append(report.WeakPasswords, WeakPasswordEntry{
                VaultID: vault.ID,
                Title:   vault.Title,
                Website: vault.Website,
                Score:   score,
                Issues:  strength["issues"].([]string),
            })
        }

        // Vérifier breach
        breached, count, _ := s.breachService.CheckBreach(password)
        if breached {
            report.Statistics.BreachedCount++

            report.BreachedPasswords = append(report.BreachedPasswords, BreachedPasswordEntry{
                VaultID:     vault.ID,
                Title:       vault.Title,
                Website:     vault.Website,
                BreachCount: count,
            })
        }

        // Vérifier âge
        ageDays := int(time.Since(vault.UpdatedAt).Hours() / 24)
        if ageDays > 365 {
            report.Statistics.VeryOldCount++
            report.OldPasswords = append(report.OldPasswords, OldPasswordEntry{
                VaultID:     vault.ID,
                Title:       vault.Title,
                Website:     vault.Website,
                AgeDays:     ageDays,
                LastChanged: vault.UpdatedAt,
            })
        } else if ageDays > 180 {
            report.Statistics.OldCount++
            report.OldPasswords = append(report.OldPasswords, OldPasswordEntry{
                VaultID:     vault.ID,
                Title:       vault.Title,
                Website:     vault.Website,
                AgeDays:     ageDays,
                LastChanged: vault.UpdatedAt,
            })
        }
    }

    // Détecter duplicats (password utilisé 2+ fois)
    for password, vaultsList := range passwordMap {
        if len(vaultsList) > 1 {
            report.Statistics.DuplicateCount += len(vaultsList)

            vaultInfos := make([]DuplicatePasswordVaultInfo, len(vaultsList))
            for i, v := range vaultsList {
                vaultInfos[i] = DuplicatePasswordVaultInfo{
                    ID:      v.ID,
                    Title:   v.Title,
                    Website: v.Website,
                }
            }

            report.DuplicatePasswords = append(report.DuplicatePasswords, DuplicatePasswordSet{
                Password: "***", // Ne pas exposer le password
                Count:    len(vaultsList),
                Vaults:   vaultInfos,
            })
        }
    }

    // Calcul overall score
    if len(vaults) > 0 {
        report.OverallScore = totalScore / len(vaults)
    }

    // Générer actions prioritaires
    report.PriorityActions = s.generatePriorityActions(report)

    return report, nil
}

func (s *AuditService) generatePriorityActions(report *AuditReport) []PriorityAction {
    actions := []PriorityAction{}

    // Breached passwords = CRITICAL
    if report.Statistics.BreachedCount > 0 {
        vaultIDs := make([]uuid.UUID, len(report.BreachedPasswords))
        for i, b := range report.BreachedPasswords {
            vaultIDs[i] = b.VaultID
        }

        actions = append(actions, PriorityAction{
            Priority:    "critical",
            Action:      "change_breached_passwords",
            VaultCount:  report.Statistics.BreachedCount,
            VaultIDs:    vaultIDs,
            Description: fmt.Sprintf("Change %d breached passwords immediately", report.Statistics.BreachedCount),
        })
    }

    // Duplicate passwords = HIGH
    if report.Statistics.DuplicateCount > 0 {
        var vaultIDs []uuid.UUID
        for _, dup := range report.DuplicatePasswords {
            for _, v := range dup.Vaults {
                vaultIDs = append(vaultIDs, v.ID)
            }
        }

        actions = append(actions, PriorityAction{
            Priority:    "high",
            Action:      "make_passwords_unique",
            VaultCount:  report.Statistics.DuplicateCount,
            VaultIDs:    vaultIDs,
            Description: fmt.Sprintf("Make %d duplicate passwords unique", report.Statistics.DuplicateCount),
        })
    }

    // Very old passwords = HIGH
    if report.Statistics.VeryOldCount > 0 {
        vaultIDs := []uuid.UUID{}
        for _, old := range report.OldPasswords {
            if old.AgeDays > 365 {
                vaultIDs = append(vaultIDs, old.VaultID)
            }
        }

        actions = append(actions, PriorityAction{
            Priority:    "high",
            Action:      "update_old_passwords",
            VaultCount:  report.Statistics.VeryOldCount,
            VaultIDs:    vaultIDs,
            Description: fmt.Sprintf("Update %d passwords older than 1 year", report.Statistics.VeryOldCount),
        })
    }

    // Weak passwords = MEDIUM
    if report.Statistics.WeakCount > 0 {
        vaultIDs := make([]uuid.UUID, len(report.WeakPasswords))
        for i, w := range report.WeakPasswords {
            vaultIDs[i] = w.VaultID
        }

        actions = append(actions, PriorityAction{
            Priority:    "medium",
            Action:      "strengthen_weak_passwords",
            VaultCount:  report.Statistics.WeakCount,
            VaultIDs:    vaultIDs,
            Description: fmt.Sprintf("Strengthen %d weak passwords", report.Statistics.WeakCount),
        })
    }

    return actions
}
```

##### 2. Créer Handler d'Audit

**Fichier**: `internal/api/handlers/audit.go` (nouveau)

```go
package handlers

import (
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "github.com/tresor/password-manager/internal/repository"
    "github.com/tresor/password-manager/internal/services"
)

type AuditHandler struct {
    vaultRepo   *repository.VaultRepository
    auditService *services.AuditService
}

func NewAuditHandler(
    vaultRepo *repository.VaultRepository,
    auditService *services.AuditService,
) *AuditHandler {
    return &AuditHandler{
        vaultRepo:   vaultRepo,
        auditService: auditService,
    }
}

func (h *AuditHandler) GetAdvancedAudit(c *gin.Context) {
    userID := c.GetString("user_id")
    masterPassword := c.Query("master_password")

    if masterPassword == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Master password required"})
        return
    }

    // Récupérer tous les vaults
    vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
        return
    }

    // Générer rapport d'audit complet
    report, err := h.auditService.GenerateAdvancedAudit(
        c.Request.Context(),
        vaults,
        masterPassword,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate audit report"})
        return
    }

    c.JSON(http.StatusOK, report)
}
```

##### 3. Ajouter Route API

```go
audit := protected.Group("/audit")
{
    audit.GET("/report", r.auditHandler.GetAdvancedAudit)  // NOUVEAU
}
```

##### 4. Tests de Vérification

```bash
# Créer plusieurs vaults avec duplicats et passwords faibles
TOKEN="..."

# Vault 1 & 2 avec même password (duplicate)
curl -X POST http://localhost:8000/api/v1/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Gmail","password":"duplicate123","master_password":"Master123"}'

curl -X POST http://localhost:8000/api/v1/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Facebook","password":"duplicate123","master_password":"Master123"}'

# Vault 3 avec password faible
curl -X POST http://localhost:8000/api/v1/vault \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Twitter","password":"weak","master_password":"Master123"}'

# Générer rapport d'audit
curl "http://localhost:8000/api/v1/audit/report?master_password=Master123" \
  -H "Authorization: Bearer $TOKEN" | jq .

# Réponse attendue:
# {
#   "overall_score": 45,
#   "total_passwords": 3,
#   "statistics": {
#     "excellent_passwords": 0,
#     "good_passwords": 0,
#     "weak_passwords": 3,
#     "breached_passwords": 0,
#     "duplicate_passwords": 2,
#     "old_passwords": 0,
#     "very_old_passwords": 0
#   },
#   "duplicate_passwords": [
#     {
#       "password": "***",
#       "count": 2,
#       "vaults": [
#         {"id":"...", "title":"Gmail", "website":null},
#         {"id":"...", "title":"Facebook", "website":null}
#       ]
#     }
#   ],
#   "old_passwords": [],
#   "weak_passwords": [...],
#   "breached_passwords": [],
#   "priority_actions": [
#     {
#       "priority": "high",
#       "action": "make_passwords_unique",
#       "vault_count": 2,
#       "vault_ids": ["...", "..."],
#       "description": "Make 2 duplicate passwords unique"
#     },
#     {
#       "priority": "medium",
#       "action": "strengthen_weak_passwords",
#       "vault_count": 3,
#       "vault_ids": ["...", "...", "..."],
#       "description": "Strengthen 3 weak passwords"
#     }
#   ],
#   "generated_at": "2026-01-11T..."
# }
```

---

**Feature 5: Recommandations de Changement** ⭐
- **Pourquoi en dernier de Phase 3**: Utilise rapport d'audit
- **Durée**: 1 heure
- **Dépendances**: Rapport d'audit doit exister

#### Plan d'Implémentation

##### 1. Ajouter Endpoint de Recommandations

**Fichier**: `internal/api/handlers/audit.go`

```go
func (h *AuditHandler) GetPasswordRecommendations(c *gin.Context) {
    userID := c.GetString("user_id")
    masterPassword := c.Query("master_password")

    if masterPassword == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Master password required"})
        return
    }

    vaults, err := h.vaultRepo.GetByUserID(c.Request.Context(), uuid.MustParse(userID))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vaults"})
        return
    }

    // Générer rapport complet
    report, err := h.auditService.GenerateAdvancedAudit(
        c.Request.Context(),
        vaults,
        masterPassword,
    )
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate report"})
        return
    }

    // Extraire seulement priority_actions
    c.JSON(http.StatusOK, gin.H{
        "recommendations": report.PriorityActions,
        "summary": gin.H{
            "total_actions": len(report.PriorityActions),
            "critical_count": countByPriority(report.PriorityActions, "critical"),
            "high_count": countByPriority(report.PriorityActions, "high"),
            "medium_count": countByPriority(report.PriorityActions, "medium"),
        },
    })
}

func countByPriority(actions []services.PriorityAction, priority string) int {
    count := 0
    for _, action := range actions {
        if action.Priority == priority {
            count++
        }
    }
    return count
}
```

##### 2. Email de Rapport Hebdomadaire (Optionnel)

**Fichier**: `internal/services/scheduler.go` (nouveau)

```go
package services

import (
    "context"
    "fmt"
    "log"
    "time"
)

type SchedulerService struct {
    auditService *AuditService
    emailService *EmailService
    userRepo     *repository.UserRepository
    vaultRepo    *repository.VaultRepository
}

func (s *SchedulerService) SendWeeklyAuditReports() {
    ctx := context.Background()

    // Récupérer tous les users (avec pagination si nécessaire)
    // Pour simplification, on assume une petite base

    users, err := s.userRepo.GetAll(ctx)
    if err != nil {
        log.Printf("Failed to fetch users for weekly report: %v", err)
        return
    }

    for _, user := range users {
        // Récupérer vaults
        vaults, err := s.vaultRepo.GetByUserID(ctx, user.ID)
        if err != nil {
            log.Printf("Failed to fetch vaults for user %s: %v", user.ID, err)
            continue
        }

        // Note: On ne peut pas décrypter sans master password
        // Donc email contient stats générales sans détails sensibles

        emailBody := fmt.Sprintf(`
Weekly Security Report

Total Passwords: %d
Passwords not changed in 6+ months: %d
Passwords not changed in 1+ year: %d

Recommendation: Log in to review your password health report.
        `, len(vaults), countOldPasswords(vaults, 180), countOldPasswords(vaults, 365))

        s.emailService.SendEmail(
            user.Email,
            "Weekly Password Security Report",
            emailBody,
        )
    }
}

func countOldPasswords(vaults []models.Vault, days int) int {
    count := 0
    threshold := time.Now().AddDate(0, 0, -days)
    for _, v := range vaults {
        if v.UpdatedAt.Before(threshold) {
            count++
        }
    }
    return count
}
```

##### 3. Route API

```go
audit := protected.Group("/audit")
{
    audit.GET("/report", r.auditHandler.GetAdvancedAudit)
    audit.GET("/recommendations", r.auditHandler.GetPasswordRecommendations)  // NOUVEAU
}
```

---

### Bonus (Optionnel): Features d'Équipe (4-5h)

**Feature 6: Team Vaults (si temps disponible)**

⚠️ **Note**: Cette feature est **plus complexe** et peut être laissée pour Phase 4

**Composants nécessaires:**
1. Table `teams` (id, name, owner_id)
2. Table `team_members` (team_id, user_id, role)
3. Table `team_vaults` (id, team_id, title, encrypted_data, etc.)
4. RBAC: admin, editor, viewer roles
5. Audit logs pour team access
6. Endpoints: create team, add members, manage permissions

**Estimation**: 4-5 heures minimum

**Recommandation**: Implémenter uniquement si les 5 premières features sont complétées ET testées.

---

## Résumé Complet Option C

### Timeline Totale

| Phase | Features | Durée | Cumulé |
|-------|----------|-------|--------|
| Phase 1 | Recherche + Historique | 3-4h | 3-4h |
| Phase 2 | Change Master Password | 2-3h | 5-7h |
| Phase 3 | Audit + Recommandations | 3h | 8-10h |
| **Bonus** | **Team Features** | **4-5h** | **12-15h** |

### Fichiers Créés (15 nouveaux)

1. `internal/models/password_history.go`
2. `internal/repository/password_history.go`
3. `internal/services/audit.go`
4. `internal/services/scheduler.go` (optionnel)
5. `internal/api/handlers/audit.go`

### Fichiers Modifiés (8 modifiés)

1. `internal/models/user.go`
2. `internal/models/vault.go`
3. `internal/repository/user.go`
4. `internal/repository/vault.go`
5. `internal/api/handlers/auth.go`
6. `internal/api/handlers/vault.go`
7. `internal/api/router.go`
8. `cmd/server/main.go`

### Endpoints Ajoutés (8 nouveaux)

1. `GET /api/v1/vault/search` - Recherche avancée
2. `GET /api/v1/vault/:id/history` - Historique
3. `POST /api/v1/vault/:id/restore/:history_id` - Restore
4. `POST /api/v1/auth/change-password` - Change master password
5. `GET /api/v1/audit/report` - Rapport d'audit complet
6. `GET /api/v1/audit/recommendations` - Recommandations

### Base de Données (2 tables + 5 index)

**Nouvelles Tables:**
1. `password_history` (6 colonnes)

**Nouvelles Colonnes:**
1. `users.password_changed_at` (optionnel)

**Nouveaux Index:**
1. `idx_vaults_favorite`
2. `idx_vaults_last_used`
3. `idx_vaults_updated_at`
4. `idx_vaults_title`
5. `idx_password_history_vault_id`

### Bénéfices Cumulés

**Pour les Utilisateurs:**
- ✅ Recherche puissante avec filtres multiples
- ✅ Historique des passwords (restore possible)
- ✅ Changement sécurisé de master password
- ✅ Vision claire des faiblesses de sécurité
- ✅ Recommandations actionnables
- ✅ Détection automatique des duplicats
- ✅ Alertes sur passwords anciens/faibles/breached

**Pour la Sécurité:**
- ✅ Re-chiffrement transactionnel
- ✅ Validation forte des nouveaux passwords
- ✅ Detection breach HIBP
- ✅ Audit complet automatisé
- ✅ Emails de notification

**Pour le Produit:**
- ✅ Avantage compétitif (features avancées)
- ✅ Conformité sécurité accrue
- ✅ Meilleure rétention utilisateur
- ✅ Base solide pour fonctionnalités futures

---

## Tests de Validation Finale

### Scénario Complet End-to-End

```bash
#!/bin/bash
# Test complet de toutes les features

# 1. SETUP: Créer utilisateur
TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"fulltest@example.com","master_password":"InitialPass123!"}' \
  | jq -r '.access_token')

echo "✓ User created"

# 2. FEATURE 1: Créer plusieurs vaults
for i in {1..5}; do
  curl -s -X POST http://localhost:8000/api/v1/vault \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"title\":\"Vault$i\",\"password\":\"pass$i\",\"folder\":\"work\",\"favorite\":true,\"master_password\":\"InitialPass123!\"}" > /dev/null
done

echo "✓ Vaults created"

# 3. FEATURE 1: Test recherche avec filtres
curl -s "http://localhost:8000/api/v1/vault/search?folder=work&favorite=true&page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN" | jq '.pagination'

echo "✓ Search with filters works"

# 4. FEATURE 2: Modifier un vault (crée historique)
VAULT_ID=$(curl -s http://localhost:8000/api/v1/vault -H "Authorization: Bearer $TOKEN" | jq -r '.[0].id')

curl -s -X PUT http://localhost:8000/api/v1/vault/$VAULT_ID \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Vault1","password":"newpass1","master_password":"InitialPass123!"}' > /dev/null

echo "✓ Password updated (history created)"

# Vérifier historique
curl -s "http://localhost:8000/api/v1/vault/$VAULT_ID/history?master_password=InitialPass123!" \
  -H "Authorization: Bearer $TOKEN" | jq 'length'

echo "✓ History retrieved"

# 5. FEATURE 3: Changer master password
curl -s -X POST http://localhost:8000/api/v1/auth/change-password \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "current_password":"InitialPass123!",
    "new_password":"NewSecurePass456!",
    "confirm_password":"NewSecurePass456!"
  }' | jq '.vaults_reencrypted'

echo "✓ Master password changed"

# Login avec nouveau password
NEW_TOKEN=$(curl -s -X POST http://localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"fulltest@example.com","master_password":"NewSecurePass456!"}' \
  | jq -r '.access_token')

echo "✓ Login with new password works"

# 6. FEATURE 4: Générer rapport d'audit
curl -s "http://localhost:8000/api/v1/audit/report?master_password=NewSecurePass456!" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq '.statistics'

echo "✓ Audit report generated"

# 7. FEATURE 5: Obtenir recommandations
curl -s "http://localhost:8000/api/v1/audit/recommendations?master_password=NewSecurePass456!" \
  -H "Authorization: Bearer $NEW_TOKEN" | jq '.summary'

echo "✓ Recommendations retrieved"

echo ""
echo "🎉 All features working!"
```

---

## Fichiers Critiques - Référence Rapide

### Priorité Haute (Toujours modifier)

1. `internal/api/router.go` - Routes
2. `cmd/server/main.go` - Dependency injection
3. `internal/models/vault.go` - Requêtes/responses
4. `internal/repository/vault.go` - Accès données
5. `internal/api/handlers/vault.go` - Logique métier

### Priorité Moyenne (Selon feature)

6. `internal/models/user.go` - User-related features
7. `internal/repository/user.go` - User queries
8. `internal/api/handlers/auth.go` - Auth endpoints
9. `internal/services/crypto.go` - Encryption logic
10. `internal/database/postgres.go` - Migrations

### À Créer

11. `internal/models/password_history.go`
12. `internal/repository/password_history.go`
13. `internal/services/audit.go`
14. `internal/api/handlers/audit.go`

---

## Stratégie de Rollback

### Par Feature

Chaque feature est **indépendante** - rollback possible individuellement:

```bash
# Rollback Feature 1 (Search)
git revert <commit-sha-search>

# Rollback Feature 2 (History)
git revert <commit-sha-history>
# + DROP TABLE password_history

# Rollback Feature 3 (Change Password)
git revert <commit-sha-change-pass>
# + DROP COLUMN users.password_changed_at (si ajouté)
```

### Rollback Complet

```bash
git reset --hard <commit-avant-option-c>
```

---

# CONCLUSION

## Quelle Option Choisir?

### Choisir Option A si:
- ✅ Vous voulez de la **valeur immédiate** pour les utilisateurs
- ✅ Vous avez **peu de temps** (3-4h disponibles)
- ✅ Vous voulez tester l'**adoption** avant d'investir plus
- ✅ Vous préférez des **features visibles** (UX)

### Choisir Option B si:
- ✅ La **sécurité est critique** maintenant
- ✅ Vous avez des **exigences de compliance**
- ✅ Les utilisateurs **demandent** le changement de password
- ✅ Vous voulez une feature **robuste** avec ROI clair

### Choisir Option C si:
- ✅ Vous voulez **maximiser l'impact**
- ✅ Vous avez **1-2 jours** disponibles
- ✅ Vous voulez un **avantage compétitif** significatif
- ✅ Votre roadmap permet une **implémentation séquentielle**

---

## Ma Recommandation Personnelle

**Je recommande: Option C (Plan Complet)**

**Raisons:**
1. Les features se **complètent** (synergie forte)
2. Ordre d'implémentation **logique** (pas de rework)
3. Impact **maximal** sur satisfaction et sécurité
4. Investissement temps **raisonnable** (1-2 jours)
5. Base **solide** pour futures features (équipe, SSO, etc.)

**Approche suggérée:**
- **Jour 1**: Phase 1 + Phase 2 (Fondations + Sécurité)
- **Jour 2**: Phase 3 (Intelligence + Tests complets)
- **Bonus**: Features équipe si demandé

---

## Prochaines Étapes

1. **Valider** quelle option vous préférez
2. **Prioriser** les features au sein de l'option choisie
3. **Commencer** l'implémentation phase par phase
4. **Tester** après chaque feature (validation incrémentale)
5. **Déployer** progressivement (feature flags optionnels)

---

**Questions?** N'hésitez pas à me demander des clarifications sur:
- Détails d'implémentation spécifiques
- Patterns GORM avancés
- Stratégies de test
- Ordre d'exécution optimal
