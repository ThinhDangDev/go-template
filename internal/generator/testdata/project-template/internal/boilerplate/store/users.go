package store

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"__MODULE_PATH__/internal/boilerplate/telemetry"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetByEmail(ctx context.Context, email string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	start := time.Now()

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id::text, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1`,
		email,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	telemetry.ObserveDBQuery("users.get_by_email", time.Since(start), err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) GetByID(ctx context.Context, id string) (*User, error) {
	start := time.Now()

	row := s.db.QueryRowContext(
		ctx,
		`SELECT id::text, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1`,
		id,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	telemetry.ObserveDBQuery("users.get_by_id", time.Since(start), err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (s *UserStore) UpsertAdmin(ctx context.Context, email, passwordHash, role string) (*User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	start := time.Now()

	row := s.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash, role, is_active)
		VALUES ($1, $2, $3, TRUE)
		ON CONFLICT (email) DO UPDATE
		SET password_hash = EXCLUDED.password_hash,
			role = EXCLUDED.role,
			is_active = TRUE,
			updated_at = NOW()
		RETURNING id::text, email, password_hash, role, is_active, created_at, updated_at`,
		email,
		passwordHash,
		role,
	)

	var user User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	telemetry.ObserveDBQuery("users.upsert_admin", time.Since(start), err)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
