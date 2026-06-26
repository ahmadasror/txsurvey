package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ahmadasror/txsurvey/internal/model"
)

// UserRepo persists creators.
type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// UpsertByGoogleSub inserts the user on first sign-in, or refreshes their
// profile on return. google_sub is the stable conflict key.
func (r *UserRepo) UpsertByGoogleSub(ctx context.Context, p model.GoogleProfile) (*model.User, error) {
	const q = `
		INSERT INTO users (google_sub, email, name, picture_url)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (google_sub) DO UPDATE
		    SET email = EXCLUDED.email,
		        name = EXCLUDED.name,
		        picture_url = EXCLUDED.picture_url,
		        updated_at = now()
		RETURNING id, google_sub, email, name, picture_url, created_at, updated_at`
	var u model.User
	err := r.pool.QueryRow(ctx, q, p.Sub, p.Email, p.Name, p.Picture).Scan(
		&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("upsert user: %w", err)
	}
	return &u, nil
}

// UpsertByGoogleSubCapped behaves like UpsertByGoogleSub but refuses to create a
// NEW creator once the table holds maxUsers rows (returning capped=true).
// Existing users (matching google_sub) always pass — the cap blocks only new
// sign-ups. The cap is enforced atomically in one statement (no count/insert
// race): the INSERT…SELECT yields no row when the cap is hit for a new sub, so
// RETURNING is empty and we read pgx.ErrNoRows. maxUsers <= 0 means unlimited.
func (r *UserRepo) UpsertByGoogleSubCapped(ctx context.Context, p model.GoogleProfile, maxUsers int) (*model.User, bool, error) {
	if maxUsers <= 0 {
		u, err := r.UpsertByGoogleSub(ctx, p)
		return u, false, err
	}
	const q = `
		INSERT INTO users (google_sub, email, name, picture_url)
		SELECT $1, $2, $3, $4
		WHERE (SELECT count(*) FROM users) < $5
		   OR EXISTS (SELECT 1 FROM users WHERE google_sub = $1)
		ON CONFLICT (google_sub) DO UPDATE
		    SET email = EXCLUDED.email,
		        name = EXCLUDED.name,
		        picture_url = EXCLUDED.picture_url,
		        updated_at = now()
		RETURNING id, google_sub, email, name, picture_url, created_at, updated_at`
	var u model.User
	err := r.pool.QueryRow(ctx, q, p.Sub, p.Email, p.Name, p.Picture, maxUsers).Scan(
		&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, true, nil // cap reached for a new account
	}
	if err != nil {
		return nil, false, fmt.Errorf("upsert user (capped): %w", err)
	}
	return &u, false, nil
}

// GetByID returns the user, or (nil, nil) when not found.
func (r *UserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	const q = `
		SELECT id, google_sub, email, name, picture_url, created_at, updated_at
		FROM users WHERE id = $1`
	var u model.User
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&u.ID, &u.GoogleSub, &u.Email, &u.Name, &u.PictureURL, &u.CreatedAt, &u.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &u, nil
}
