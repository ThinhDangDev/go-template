package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"__MODULE_PATH__/internal/boilerplate/telemetry"
	"__MODULE_PATH__/internal/domain"
	"__MODULE_PATH__/internal/domain/entity"

	"github.com/lib/pq"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *userRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	start := time.Now()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id::text, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1`,
		email,
	)

	var user entity.User
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
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entity.User, error) {
	start := time.Now()

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id::text, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1`,
		id,
	)

	var user entity.User
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
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, email, passwordHash, role string) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	start := time.Now()

	row := r.db.QueryRowContext(
		ctx,
		`INSERT INTO users (email, password_hash, role, is_active)
		VALUES ($1, $2, $3, TRUE)
		RETURNING id::text, email, password_hash, role, is_active, created_at, updated_at`,
		email,
		passwordHash,
		role,
	)

	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	telemetry.ObserveDBQuery("users.create", time.Since(start), err)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return nil, domain.ErrEmailAlreadyExists
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) List(ctx context.Context) ([]*entity.User, error) {
	start := time.Now()

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id::text, email, password_hash, role, is_active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC, email ASC`,
	)
	if err != nil {
		telemetry.ObserveDBQuery("users.list", time.Since(start), err)
		return nil, err
	}
	defer rows.Close()

	users := make([]*entity.User, 0)
	for rows.Next() {
		var user entity.User
		if err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.PasswordHash,
			&user.Role,
			&user.IsActive,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			telemetry.ObserveDBQuery("users.list", time.Since(start), err)
			return nil, err
		}
		users = append(users, &user)
	}

	err = rows.Err()
	telemetry.ObserveDBQuery("users.list", time.Since(start), err)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) UpdateAccess(ctx context.Context, id, role string, isActive bool) (*entity.User, error) {
	start := time.Now()

	row := r.db.QueryRowContext(
		ctx,
		`UPDATE users
		SET role = $2,
			is_active = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id::text, email, password_hash, role, is_active, created_at, updated_at`,
		id,
		role,
		isActive,
	)

	var user entity.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	telemetry.ObserveDBQuery("users.update_access", time.Since(start), err)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) UpsertAdmin(ctx context.Context, email, passwordHash, role string) (*entity.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	start := time.Now()

	row := r.db.QueryRowContext(
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

	var user entity.User
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
