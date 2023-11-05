package repository

import (
	"context"
	"database/sql"
	"testcontainers/internal/post"
	"time"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{
		db: db,
	}
}

func (r *PostRepository) Create(ctx context.Context, p *post.Post) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	q := "INSERT INTO posts (id, title, content) VALUES ($1, $2, $3);"
	_, err := r.db.ExecContext(
		ctx,
		q,
		p.ID,
		p.Title,
		p.Content,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostRepository) Find(ctx context.Context, id string) (*post.Post, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*5)
	defer cancel()

	var p post.Post

	q := "SELECT * FROM posts WHERE id = $1;"
	err := r.db.QueryRowContext(
		ctx,
		q,
		id,
	).Scan(
		&p.ID,
		&p.Title,
		&p.Content,
	)
	if err != nil {
		return nil, err
	}

	return &p, nil
}
