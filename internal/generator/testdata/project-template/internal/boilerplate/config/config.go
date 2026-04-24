package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	ServiceName      string
	Environment      string
	Version          string
	HTTPHost         string
	HTTPPort         int
	GRPCHost         string
	GRPCPort         int
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	ShutdownTimeout  time.Duration
	DatabaseURL      string
	PostgresHost     string
	PostgresPort     string
	PostgresUser     string
	PostgresPassword string
	PostgresDB       string
	PostgresSSLMode  string
	JWTSecret        string
	JWTIssuer        string
	JWTAccessTTL     time.Duration
	OTLPEndpoint     string
	OTLPInsecure     bool
	LogLevel         string
	LogEnableConsole bool
	LogEnableFile    bool
	LogFile          string
	MigrationsDir    string
	CasbinModelPath  string
	SwaggerJSONPath  string
	AdminEmail       string
	AdminPassword    string
	AdminRole        string
}

func Load() (Config, error) {
	_ = godotenv.Load()

	cfg := Config{
		ServiceName:      envOrDefault("SERVICE_NAME", "__PROJECT_NAME__"),
		Environment:      envOrDefault("APP_ENV", "local"),
		Version:          envOrDefault("APP_VERSION", "dev"),
		HTTPHost:         envOrDefault("HTTP_HOST", "0.0.0.0"),
		HTTPPort:         envAsInt("HTTP_PORT", 8080),
		GRPCHost:         envOrDefault("GRPC_HOST", "0.0.0.0"),
		GRPCPort:         envAsInt("GRPC_PORT", 9090),
		HTTPReadTimeout:  envAsDuration("HTTP_READ_TIMEOUT", 15*time.Second),
		HTTPWriteTimeout: envAsDuration("HTTP_WRITE_TIMEOUT", 15*time.Second),
		ShutdownTimeout:  envAsDuration("SHUTDOWN_TIMEOUT", 10*time.Second),
		PostgresHost:     envOrDefault("POSTGRES_HOST", "localhost"),
		PostgresPort:     envOrDefault("POSTGRES_PORT", "5432"),
		PostgresUser:     envOrDefault("POSTGRES_USER", "postgres"),
		PostgresPassword: envOrDefault("POSTGRES_PASSWORD", "postgres"),
		PostgresDB:       envOrDefault("POSTGRES_DB", "__PROJECT_NAME_SNAKE__"),
		PostgresSSLMode:  envOrDefault("POSTGRES_SSLMODE", "disable"),
		JWTSecret:        strings.TrimSpace(os.Getenv("JWT_SECRET")),
		JWTIssuer:        envOrDefault("JWT_ISSUER", "__PROJECT_NAME__"),
		JWTAccessTTL:     envAsDuration("JWT_ACCESS_TTL", 1*time.Hour),
		OTLPEndpoint:     os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"),
		OTLPInsecure:     envAsBool("OTEL_EXPORTER_OTLP_INSECURE", true),
		LogLevel:         envOrDefault("LOG_LEVEL", "info"),
		LogEnableConsole: envAsBool("LOG_ENABLE_CONSOLE", true),
		LogEnableFile:    envAsBool("LOG_ENABLE_FILE", true),
		LogFile:          envOrDefault("LOG_FILE", "logs/__PROJECT_NAME__.log"),
		MigrationsDir:    envOrDefault("MIGRATIONS_DIR", "migrations/sql"),
		CasbinModelPath:  envOrDefault("CASBIN_MODEL_PATH", "configs/rbac_model.conf"),
		SwaggerJSONPath:  envOrDefault("SWAGGER_JSON_PATH", "internal/docs/api.swagger.json"),
		AdminEmail:       envOrDefault("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword:    strings.TrimSpace(os.Getenv("ADMIN_PASSWORD")),
		AdminRole:        envOrDefault("ADMIN_ROLE", "admin"),
	}

	cfg.DatabaseURL = os.Getenv("DATABASE_URL")
	if cfg.DatabaseURL == "" {
		u := &url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort(cfg.PostgresHost, cfg.PostgresPort),
			Path:   cfg.PostgresDB,
			RawQuery: url.Values{
				"sslmode": []string{cfg.PostgresSSLMode},
			}.Encode(),
		}
		u.User = url.UserPassword(cfg.PostgresUser, cfg.PostgresPassword)
		cfg.DatabaseURL = u.String()
	}

	return cfg, nil
}

func (c Config) HTTPAddress() string {
	return fmt.Sprintf("%s:%d", c.HTTPHost, c.HTTPPort)
}

func (c Config) GRPCAddress() string {
	return fmt.Sprintf("%s:%d", c.GRPCHost, c.GRPCPort)
}

func (c Config) ResolvedMigrationsDir() string {
	return resolvePath(c.MigrationsDir)
}

func (c Config) ResolvedCasbinModelPath() string {
	return resolvePath(c.CasbinModelPath)
}

func (c Config) ResolvedSwaggerJSONPath() string {
	return resolvePath(c.SwaggerJSONPath)
}

func resolvePath(path string) string {
	if path == "" {
		return path
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}

	return abs
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func envAsBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envAsInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}

func envAsDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}

	return parsed
}
