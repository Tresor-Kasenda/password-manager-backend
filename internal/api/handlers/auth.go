package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/tresor/password-manager/internal/config"
	"github.com/tresor/password-manager/internal/models"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

type AuthHandler struct {
	userRepo      *repository.UserRepository
	cryptoService *services.CryptoService
	jwtSecret     string
	jwtExpire     int
}

func NewAuthHandler(
	userRepo *repository.UserRepository,
	cryptoService *services.CryptoService,
	cfg *config.JWTConfig,
) *AuthHandler {
	return &AuthHandler{
		userRepo:      userRepo,
		cryptoService: cryptoService,
		jwtSecret:     cfg.Secret,
		jwtExpire:     cfg.ExpireTime,
	}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existingUser, _ := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	hash, salt, err := h.cryptoService.HashPassword(req.MasterPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	user := &models.User{
		ID:                 uuid.New(),
		Email:              req.Email,
		MasterPasswordHash: hash,
		Salt:               salt,
		TwoFactorEnabled:   false,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := h.userRepo.Create(c.Request.Context(), user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "User created successfully"})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userRepo.GetByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !h.cryptoService.VerifyPassword(req.MasterPassword, user.MasterPasswordHash, user.Salt) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := h.generateToken(user.ID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		AccessToken: token,
		TokenType:   "bearer",
		User:        *user,
	})
}

func (h *AuthHandler) generateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(time.Hour * time.Duration(h.jwtExpire)).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.jwtSecret))
}
