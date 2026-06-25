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
