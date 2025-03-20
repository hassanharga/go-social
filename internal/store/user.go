package store

import (
	"context"
	"database/sql"
)

type UserStore struct {
	db *sql.DB
}

func (p *UserStore) Create(ctx context.Context) error {
	// Implementation for creating a post
	return nil
}
