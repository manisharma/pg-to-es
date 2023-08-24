package mock

import (
	"context"
	"fmt"
	"pg-to-es/internal/model"
	"strings"
)

type Elastic struct {
	documents []model.User
}

func NewElastic(documents []model.User) *Elastic {
	return &Elastic{documents: documents}
}

func (e *Elastic) Create(ctx context.Context, index string, id int, doc model.User) error {
	e.documents = append(e.documents, doc)
	return nil
}

func (e *Elastic) GetByProjectId(ctx context.Context, index string, projectId int) ([]model.User, error) {
	var res []model.User
	for _, document := range e.documents {
		for _, project := range document.Projects {
			if project.ID == projectId {
				res = append(res, document)
			}
		}
	}
	return res, nil
}

func (e *Elastic) GetByHashTagId(ctx context.Context, index string, hashTagId int) ([]model.User, error) {
	return nil, nil
}

func (e *Elastic) GetByUserId(ctx context.Context, index string, userId int) (*model.User, error) {
	return nil, nil
}

func (e *Elastic) RemoveProject(ctx context.Context, index string, projectId int) error {
	return nil
}

func (e *Elastic) RemoveHashtag(ctx context.Context, index string, hashtagId int) error {
	return nil
}

func (e *Elastic) Update(ctx context.Context, index string, id int, user model.User) error {
	return nil
}

func (e *Elastic) Delete(ctx context.Context, index string, id int) error {
	return nil
}

func (e *Elastic) SearchByUser(ctx context.Context, index string, userID int) (*model.User, error) {
	for _, document := range e.documents {
		if document.ID == userID {
			return &document, nil
		}
	}
	return nil, fmt.Errorf("%s", "Not found")
}

func (e *Elastic) SearchByHashtags(ctx context.Context, index string, hashtag string) ([]model.User, error) {
	return nil, nil
}

func (e *Elastic) FuzzySearchProjects(ctx context.Context, index string, query string) ([]model.FuzzyResult, error) {
	var res []model.FuzzyResult
	for _, document := range e.documents {
		for _, project := range document.Projects {
			if strings.Contains(project.Description, query) || strings.Contains(project.Slug, query) {
				hashtags := []string{}
				for _, hashtag := range project.Hashtags {
					hashtags = append(hashtags, hashtag.Name)
				}
				res = append(res, model.FuzzyResult{
					Hashtags: hashtags,
					User: model.FuzzyUser{
						ID:        document.ID,
						Name:      document.Name,
						CreatedAt: document.CreatedAt,
					},
				})
			}
		}
	}
	return res, nil
}
