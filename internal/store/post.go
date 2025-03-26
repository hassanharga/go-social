package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	Version   int       `json:"version"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
	Comments  []Comment `json:"comments"`
	Users     User      `json:"user"`
}

type PostWithMetadata struct {
	Post
	CommentsCount int64 `json:"comments_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `
		INSERT INTO posts (title, content, user_id, tags)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(
		ctx,
		query,
		post.Title,
		post.Content,
		post.UserID,
		pq.Array(post.Tags),
	).Scan(
		&post.ID,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil

}

func (s *PostStore) GetById(ctx context.Context, id int64) (*Post, error) {
	query := `
		SELECT id, user_id, title, content, tags, version, created_at, updated_at 
		FROM  posts
		WHERE id = $1
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	var post Post
	if err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.UserID,
		&post.Title,
		&post.Content,
		pq.Array(&post.Tags),
		&post.Version,
		&post.CreatedAt,
		&post.UpdatedAt,
	); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}

func (s *PostStore) Delete(ctx context.Context, postId int64) error {
	query := `DELETE FROM posts WHERE id = $1`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	res, err := s.db.ExecContext(ctx, query, postId)
	if err != nil {
		return err
	}

	if affectedRows, err := res.RowsAffected(); err != nil {
		return err
	} else if affectedRows == 0 {
		return ErrNotFound
	}

	return nil
}

func (s *PostStore) Update(ctx context.Context, post *Post) error {
	query := `
		UPDATE posts 
		SET title = $1, content = $2 , version = version + 1
		WHERE id = $3 AND version = $4
		RETURNING version
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	err := s.db.QueryRowContext(ctx, query, post.Title, post.Content, post.ID, post.Version).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrNotFound
		default:
			return err
		}
	}

	return nil
}

func (s *PostStore) GetFeed(ctx context.Context, userId int64) ([]*PostWithMetadata, error) {

	query := `
		SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, u.username, COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN users u ON p.user_id = u.id
		JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
		WHERE f.user_id = $1 OR p.user_id = $1
		GROUP BY p.id, u.username
		ORDER BY p.created_at DESC
	`
	// query := `
	// 	SELECT p.id, p.user_id, p.title, p.content, p.created_at, p.version, p.tags, u.username, COUNT(c.id) AS comments_count
	// 	FROM posts p
	// 	LEFT JOIN comments c ON c.post_id = p.id
	// 	LEFT JOIN users u ON p.user_id = u.id
	// 	JOIN followers f ON f.follower_id = p.user_id OR p.user_id = $1
	// 	WHERE
	// 		f.user_id = $1 AND
	// 		(p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%') AND
	// 		(p.tags @> $5 OR $5 = '{}')
	// 	GROUP BY p.id, u.username
	// 	ORDER BY p.created_at ` + fq.Sort + `
	// 	LIMIT $2 OFFSET $3
	// `
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []*PostWithMetadata

	for rows.Next() {
		var post PostWithMetadata
		if err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			&post.CreatedAt,
			&post.Version,
			pq.Array(&post.Tags),
			&post.Users.Username,
			&post.CommentsCount,
		); err != nil {
			return nil, err
		}
		posts = append(posts, &post)
	}

	return posts, nil
}
