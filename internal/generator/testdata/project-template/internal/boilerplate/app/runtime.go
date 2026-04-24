package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"__MODULE_PATH__/internal/boilerplate/auth"
	"__MODULE_PATH__/internal/boilerplate/config"
	"__MODULE_PATH__/internal/boilerplate/telemetry"
	repositorycontract "__MODULE_PATH__/internal/domain/interface/repository"
	usecasecontract "__MODULE_PATH__/internal/domain/interface/usecase"
	"__MODULE_PATH__/internal/repository"
	"__MODULE_PATH__/internal/usecase"

	_ "github.com/lib/pq"
)

type Runtime struct {
	Config        config.Config
	Logger        *slog.Logger
	DB            *sql.DB
	Users         repositorycontract.UserRepository
	Tokens        *auth.TokenManager
	Authorizer    *auth.Authorizer
	AuthUseCase   usecasecontract.AuthUseCase
	AdminUseCase  usecasecontract.AdminUseCase
	SystemUseCase usecasecontract.SystemUseCase
	traceShutdown func(context.Context) error
}

func Bootstrap(ctx context.Context) (*Runtime, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	logger, err := telemetry.NewLogger(cfg)
	if err != nil {
		return nil, err
	}

	traceShutdown, err := telemetry.InitTracing(ctx, cfg.ServiceName, cfg.Version, cfg.OTLPEndpoint, cfg.OTLPInsecure)
	if err != nil {
		return nil, err
	}

	db, err := OpenDB(ctx, cfg)
	if err != nil {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = traceShutdown(shutdownCtx)
		return nil, err
	}

	if cfg.JWTSecret == "" {
		_ = db.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = traceShutdown(shutdownCtx)
		return nil, fmt.Errorf("JWT_SECRET is required for serve/seed workflows")
	}

	if err := EnsureRequiredSchema(ctx, db); err != nil {
		_ = db.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = traceShutdown(shutdownCtx)
		return nil, err
	}

	authorizer, err := auth.NewAuthorizer(db, cfg.ResolvedCasbinModelPath())
	if err != nil {
		_ = db.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = traceShutdown(shutdownCtx)
		return nil, err
	}

	userRepository := repository.NewUserRepository(db)
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, cfg.JWTIssuer, cfg.JWTAccessTTL)
	authUseCase := usecase.NewAuthUseCase(userRepository, tokenManager, logger, cfg.JWTAccessTTL)
	adminUseCase := usecase.NewAdminUseCase(userRepository)
	systemUseCase := usecase.NewSystemUseCase()

	telemetry.MarkAppInfo(cfg.ServiceName, cfg.Version, cfg.Environment)

	return &Runtime{
		Config:        cfg,
		Logger:        logger,
		DB:            db,
		Users:         userRepository,
		Tokens:        tokenManager,
		Authorizer:    authorizer,
		AuthUseCase:   authUseCase,
		AdminUseCase:  adminUseCase,
		SystemUseCase: systemUseCase,
		traceShutdown: traceShutdown,
	}, nil
}

func OpenDB(ctx context.Context, cfg config.Config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, err
	}

	return db, nil
}

func EnsureRequiredSchema(ctx context.Context, db *sql.DB) error {
	requiredTables := []string{
		"public.users",
		"public.casbin_rule",
	}

	for _, tableName := range requiredTables {
		var resolved sql.NullString
		if err := db.QueryRowContext(ctx, "SELECT to_regclass($1)", tableName).Scan(&resolved); err != nil {
			return err
		}
		if !resolved.Valid || resolved.String == "" {
			return fmt.Errorf("required table %s is missing; run `go run ./cmd/main.go migrate up` first", tableName)
		}
	}

	return nil
}

func (r *Runtime) Close(ctx context.Context) error {
	var closeErr error

	if r.DB != nil {
		closeErr = errors.Join(closeErr, r.DB.Close())
	}

	if r.traceShutdown != nil {
		closeErr = errors.Join(closeErr, r.traceShutdown(ctx))
	}

	return closeErr
}
