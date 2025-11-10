package app

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/templui/goilerplate/internal/config"
	"github.com/templui/goilerplate/internal/db"
	"github.com/templui/goilerplate/internal/repository"
	"github.com/templui/goilerplate/internal/service"
	"github.com/templui/goilerplate/internal/service/payment"
	"github.com/templui/goilerplate/internal/storage"
)

type App struct {
	Cfg                 *config.Config
	DB                  *sqlx.DB
	AuthService         *service.AuthService
	UserService         *service.UserService
	ProfileService      *service.ProfileService
	EmailService        *service.EmailService
	FileService         *service.FileService
	SubscriptionService *service.SubscriptionService
	PaymentService      payment.Provider
	GoalService         *service.GoalService
	BlogService         *service.BlogService
	DocsService         *service.DocsService
	LegalService        *service.LegalService
}

func New(cfg *config.Config) (*App, error) {
	// Initialize database
	database, err := db.Init(cfg.DBDriver, cfg.DBConnection)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %v", err)
	}

	// Run database migrations
	err = db.RunMigrations(database.DB, cfg.DBDriver)
	if err != nil {
		return nil, fmt.Errorf("failed to run migrations: %v", err)
	}

	// Repositories
	userRepository := repository.NewUserRepository(database)
	profileRepository := repository.NewProfileRepository(database)
	tokenRepository := repository.NewTokenRepository(database)
	fileRepository := repository.NewFileRepository(database)
	subscriptionRepository := repository.NewSubscriptionRepository(database)
	goalRepository := repository.NewGoalRepository(database)
	goalEntryRepository := repository.NewGoalEntryRepository(database)

	// Storage
	fileStorage, err := storage.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %v", err)
	}

	// Services
	emailService := service.NewEmailService(
		cfg.ResendAPIKey,
		cfg.EmailFrom,
		cfg.ResendAudienceID,
		cfg.AppURL,
		cfg.AppName,
		cfg.IsDevelopment(),
	)
	fileService := service.NewFileService(fileRepository, fileStorage)
	subscriptionService := service.NewSubscriptionService(subscriptionRepository)

	// Initialize payment provider based on config
	paymentProvider, err := payment.NewProvider(cfg, subscriptionService)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize payment provider: %v", err)
	}

	goalService := service.NewGoalService(goalRepository, goalEntryRepository, fileRepository, subscriptionService)
	authService := service.NewAuthService(
		userRepository,
		profileRepository,
		tokenRepository,
		subscriptionService,
		emailService,
		cfg.JWTSecret,
		cfg.IsProduction(),
		cfg.JWTExpiry,
		cfg.TokenEmailVerifyExpiry,
		cfg.TokenPasswordResetExpiry,
		cfg.TokenEmailChangeExpiry,
		cfg.TokenMagicLinkExpiry,
	)
	userService := service.NewUserService(userRepository, profileRepository, fileService, emailService, subscriptionService)
	profileService := service.NewProfileService(profileRepository)
	blogService := service.NewBlogService(cfg.ContentPath)
	docsService := service.NewDocsService(cfg.ContentPath)
	legalService := service.NewLegalService(cfg.ContentPath)

	return &App{
		Cfg:                 cfg,
		DB:                  database,
		AuthService:         authService,
		UserService:         userService,
		ProfileService:      profileService,
		EmailService:        emailService,
		FileService:         fileService,
		SubscriptionService: subscriptionService,
		PaymentService:      paymentProvider,
		GoalService:         goalService,
		BlogService:         blogService,
		DocsService:         docsService,
		LegalService:        legalService,
	}, nil
}

func (a *App) Close() error {
	if a.DB != nil {
		return a.DB.Close()
	}
	return nil
}
