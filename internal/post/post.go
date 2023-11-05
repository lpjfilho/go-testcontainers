package post

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var ErrTitleNotFound = errors.New("title not found")
var ErrContentNotFound = errors.New("content not found")

type Post struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

func NewPost(title, content string) (*Post, error) {
	a := &Post{
		ID:      uuid.New().String(),
		Title:   title,
		Content: content,
	}
	if err := a.Validate(); err != nil {
		return nil, err
	}

	return a, nil
}

func (p *Post) Validate() error {
	if p.Title == "" {
		return ErrTitleNotFound
	}

	if p.Content == "" {
		return ErrContentNotFound
	}

	return nil
}

type PostServiceInterface interface {
	Create(ctx context.Context, p *Post) error
	Find(ctx context.Context, id string) (*Post, error)
}

type PostReader interface {
	Find(ctx context.Context, id string) (*Post, error)
}

type PostWriter interface {
	Create(ctx context.Context, p *Post) error
}

type PostRepositoryInterface interface {
	PostReader
	PostWriter
}
