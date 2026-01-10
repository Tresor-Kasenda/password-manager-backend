package api

import (
	"github.com/gin-gonic/gin"
	"github.com/tresor/password-manager/internal/api/handlers"
	"github.com/tresor/password-manager/internal/api/middleware"
	"github.com/tresor/password-manager/internal/config"
)

type Router struct {
	authHandler    *handlers.AuthHandler
	vaultHandler   *handlers.VaultHandler
	sharingHandler *handlers.SharingHandler
	healthHandler  *handlers.HealthHandler
	twoFAHandler   *handlers.TwoFAHandler
	importHandler  *handlers.ImportHandler
	jwtSecret      string
}

func NewRouter(
	authHandler *handlers.AuthHandler,
	vaultHandler *handlers.VaultHandler,
	sharingHandler *handlers.SharingHandler,
	healthHandler *handlers.HealthHandler,
	twoFAHandler *handlers.TwoFAHandler,
	importHandler *handlers.ImportHandler,
	cfg *config.Config,
) *Router {
	return &Router{
		authHandler:    authHandler,
		vaultHandler:   vaultHandler,
		sharingHandler: sharingHandler,
		healthHandler:  healthHandler,
		twoFAHandler:   twoFAHandler,
		importHandler:  importHandler,
		jwtSecret:      cfg.JWT.Secret,
	}
}

func (r *Router) Setup() *gin.Engine {
	router := gin.Default()

	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.RateLimitMiddleware(100))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	v1 := router.Group("/api/v1")
	{
		auth := v1.Group("/auth")
		{
			auth.POST("/register", r.authHandler.Register)
			auth.POST("/login", middleware.RateLimitMiddleware(5), r.authHandler.Login)
			auth.POST("/request-deletion", r.authHandler.RequestAccountDeletion)
		}

		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(r.jwtSecret))
		{
			vault := protected.Group("/vault")
			{
				vault.POST("", r.vaultHandler.CreateVault)
				vault.GET("", r.vaultHandler.GetVaults)
				vault.GET("/:id", r.vaultHandler.GetVault)
				vault.PUT("/:id", r.vaultHandler.UpdateVault)
				vault.DELETE("/:id", r.vaultHandler.DeleteVault)
				vault.POST("/generate-password", r.vaultHandler.GeneratePassword)
				vault.POST("/scan-all", r.healthHandler.ScanAllPasswords)
			}

			health := protected.Group("/health")
			{
				health.GET("/report", r.healthHandler.GetHealthReport)
				health.POST("/analyze", r.healthHandler.AnalyzePassword)
			}

			password := protected.Group("/password")
			{
				password.POST("/check-breach", r.healthHandler.CheckPasswordBreach)
			}

			share := protected.Group("/share")
			{
				share.POST("", r.sharingHandler.SharePassword)
			}

			shared := protected.Group("/shared")
			{
				shared.GET("", r.sharingHandler.ListShares)
				shared.GET("/:token", r.sharingHandler.GetSharedPassword)
				shared.POST("/:token/revoke", r.sharingHandler.RevokeShare)
			}

			twofa := protected.Group("/2fa")
			{
				twofa.POST("/enable", r.twoFAHandler.Enable2FA)
				twofa.POST("/verify-and-enable", r.twoFAHandler.VerifyAndEnable)
				twofa.POST("/verify", r.twoFAHandler.Verify2FA)
				twofa.POST("/disable", r.twoFAHandler.Disable2FA)
			}

			importRoutes := protected.Group("/import")
			{
				importRoutes.POST("/upload", r.importHandler.UploadFile)
				importRoutes.POST("/confirm/:session_id", r.importHandler.ConfirmImport)
				importRoutes.GET("/supported-formats", r.importHandler.GetSupportedFormats)
			}
		}
	}

	return router
}
