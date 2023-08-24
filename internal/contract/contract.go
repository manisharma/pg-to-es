package contract

import (
	"context"
	"pg-to-es/internal/model"
)

type Elastic interface {
	Create(ctx context.Context, index string, id int, doc model.User) error
	GetByProjectId(ctx context.Context, index string, projectId int) ([]model.User, error)
	GetByHashTagId(ctx context.Context, index string, hashTagId int) ([]model.User, error)
	GetByUserId(ctx context.Context, index string, userId int) (*model.User, error)
	RemoveProject(ctx context.Context, index string, projectId int) error
	RemoveHashtag(ctx context.Context, index string, hashtagId int) error
	Update(ctx context.Context, index string, id int, user model.User) error
	Delete(ctx context.Context, index string, id int) error
	SearchByUser(ctx context.Context, index string, userID int) ([]model.SearchResult, error)
	SearchByHashtags(ctx context.Context, index string, hashtag string) ([]model.SearchResult, error)
	FuzzySearchProjects(ctx context.Context, index string, query string) ([]model.SearchResult, error)
}

type DbListener interface {
	Start(ctx context.Context) (<-chan string, error)
	Stop()
}
