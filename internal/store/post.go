package store

import (
	"context"
	"database/sql"
)

type PostStore struct {
	db *sql.DB
}

func (p *PostStore) Create(ctx context.Context) error {
	// Implementation for creating a post
	return nil
}
