package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tresor/password-manager/internal/api"
	"github.com/tresor/password-manager/internal/api/handlers"
	"github.com/tresor/password-manager/internal/config"
	"github.com/tresor/password-manager/internal/database"
	"github.com/tresor/password-manager/internal/repository"
	"github.com/tresor/password-manager/internal/services"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	db, err := database.NewPostgresDB(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db)
	vaultRepo := repository.NewVaultRepository(db)
	shareRepo := repository.NewShareRepository(db)

	cryptoService := services.NewCryptoService()
	emailService := services.NewEmailService(&cfg.Email)
	breachService := services.NewBreachService(cfg.HIBP.APIKey)
	passwordHealthService := services.NewPasswordHealthService()
	importService := services.NewImportService()

	authHandler := handlers.NewAuthHandler(userRepo, cryptoService, emailService, &cfg.JWT)
	vaultHandler := handlers.NewVaultHandler(vaultRepo, cryptoService)
	sharingHandler := handlers.NewSharingHandler(shareRepo, vaultRepo, userRepo, cryptoService, emailService)
	healthHandler := handlers.NewHealthHandler(vaultRepo, cryptoService, passwordHealthService, breachService)
	twoFAHandler := handlers.NewTwoFAHandler(userRepo, cryptoService)
	importHandler := handlers.NewImportHandler(vaultRepo, importService, cryptoService)

	router := api.NewRouter(
		authHandler,
		vaultHandler,
		sharingHandler,
		healthHandler,
		twoFAHandler,
		importHandler,
		cfg,
	)

	engine := router.Setup()

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Server starting on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
