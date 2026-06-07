package user

import (
	"context"
	"tms-platform/internal/model"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateWithPassword(ctx context.Context, name, email, passwordHash string) (*model.User, error) {
	const q = `
		INSERT INTO users (id, name, email, role, password_hash, created_at)
		VALUES ($1, $2, $3, $4, $5, now())
		RETURNING id, name, email, role, password_hash, created_at`

	u := model.User{}
	err := r.db.QueryRowxContext(ctx, q, uuid.New(), name, email, "viewer", passwordHash).StructScan(&u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	const q = `SELECT id, name, email, role, password_hash, created_at FROM users WHERE email = $1`

	u := model.User{}
	if err := r.db.GetContext(ctx, &u, q, email); err != nil {
		return nil, err
	}
	return &u, nil
}
