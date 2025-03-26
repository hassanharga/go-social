package store

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
	ErrConflict = errors.New("already exists")
)

const (
	QueryTimeoutDuration = time.Second * 5
)

type Storage struct {
	Posts interface {
		Create(context.Context, *Post) error
		GetById(context.Context, int64) (*Post, error)
		Delete(context.Context, int64) error
		Update(context.Context, *Post) error
		GetFeed(context.Context, int64) ([]*PostWithMetadata, error)
	}
	Comments interface {
		Create(context.Context, *Comment) error
		GetByPostId(context.Context, int64) ([]Comment, error)
	}
	Users interface {
		Create(context.Context, *User) error
		GetById(context.Context, int64) (*User, error)
	}
	Followers interface {
		Follow(ctx context.Context, followerId int64, userId int64) error
		Unfollow(ctx context.Context, followerId int64, userId int64) error
	}
}

func NewStorage(db *sql.DB) Storage {
	return Storage{
		Posts:     &PostStore{db},
		Comments:  &CommentStore{db},
		Users:     &UserStore{db},
		Followers: &FollowerStore{db},
	}
}
