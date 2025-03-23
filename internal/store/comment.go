package store

import (
	"context"
	"database/sql"
)

type Comment struct {
	ID        int64  `json:"id"`
	Content   string `json:"content"`
	UserID    int64  `json:"user_id"`
	PostID    int64  `json:"post_id"`
	CreatedAt string `json:"created_at"`
	User      User   `json:"user"`
}

type CommentStore struct {
	db *sql.DB
}

func (s *CommentStore) GetByPostId(ctx context.Context, postId int64) ([]Comment, error) {
	query := `
		SELECT c.id, c.content, c.user_id, c.post_id, c.created_at, users.username, users.id
		FROM comments c
		JOIN users ON c.user_id = users.id
		WHERE c.post_id = $1
		ORDER BY c.created_at DESC
	`

	rows, err := s.db.QueryContext(ctx, query, postId)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	comments := []Comment{}
	for rows.Next() {
		comment := Comment{}
		comment.User = User{}
		// Scan the result into the comment struct
		err := rows.Scan(
			&comment.ID,
			&comment.Content,
			&comment.UserID,
			&comment.PostID,
			&comment.CreatedAt,
			&comment.User.Username,
			&comment.User.ID,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}
	return comments, nil
}
