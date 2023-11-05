package post

import "context"

type PostService struct {
	repository PostRepositoryInterface
}

func NewPostService(repository PostRepositoryInterface) *PostService {
	return &PostService{
		repository: repository,
	}
}

func (s *PostService) Create(ctx context.Context, p *Post) error {
	// here would be the code to process the post, etc.

	err := s.repository.Create(ctx, p)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) Find(ctx context.Context, id string) (*Post, error) {
	res, err := s.repository.Find(ctx, id)

	if err != nil {
		return nil, err
	}

	return res, nil
}
