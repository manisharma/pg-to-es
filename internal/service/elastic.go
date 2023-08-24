package service

import (
	"context"
	"encoding/json"
	"fmt"
	"pg-to-es/internal/config"
	"pg-to-es/internal/model"

	"github.com/olivere/elastic/v7"
)

type Elastic struct {
	c *elastic.Client
}

func NewElastic(cfg config.Es) (*Elastic, error) {
	client, err := elastic.NewClient(
		elastic.SetURL(cfg.Host),
		elastic.SetHealthcheck(false),
		elastic.SetSniff(false))
	if err != nil {
		return nil, err
	}
	return &Elastic{client}, nil
}

// Function to create a document
func (c *Elastic) Create(ctx context.Context, index string, id int, doc model.User) error {
	_, err := c.c.Index().
		Index(index).
		Type("_doc").
		Id(fmt.Sprintf("%d", id)).
		BodyJson(doc).
		Do(ctx)
	return err
}

func (c *Elastic) GetByProjectId(ctx context.Context, index string, projectId int) ([]model.User, error) {
	var (
		query         *elastic.TermQuery
		searchService *elastic.SearchService
	)
	if projectId != 0 {
		query = elastic.NewTermQuery("projects.id", projectId)
		searchService = c.c.Search().Index(index).Query(query)
	} else {
		searchService = c.c.Search().Index(index)
	}
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, err
	}
	var results []model.User
	for _, hit := range searchResult.Hits.Hits {
		var result model.User
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (c *Elastic) GetByHashTagId(ctx context.Context, index string, hashTagId int) ([]model.User, error) {
	query := elastic.NewTermQuery("projects.hashtags.id", hashTagId)
	searchService := c.c.Search().Index(index).Query(query)
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, err
	}
	var results []model.User
	for _, hit := range searchResult.Hits.Hits {
		var result model.User
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Function to get a document
func (c *Elastic) GetByUserId(ctx context.Context, index string, userId int) (*model.User, error) {
	doc, err := c.c.Get().
		Index(index).
		Type("_doc").
		Id(fmt.Sprintf("%d", userId)).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	if doc.Found {
		var u model.User
		err = json.Unmarshal(doc.Source, &u)
		if err != nil {
			return nil, err
		}
		return &u, nil
	}
	return nil, fmt.Errorf("not found")
}

func (c *Elastic) RemoveProject(ctx context.Context, index string, projectId int) error {
	documents, err := c.GetByProjectId(ctx, index, projectId)
	if err != nil {
		return err
	}
	for _, document := range documents {
		remainigProjects := []model.Project{}
		for _, project := range document.Projects {
			if project.ID != projectId {
				remainigProjects = append(remainigProjects, project)
			}
		}
		document.Projects = remainigProjects
		err = c.Update(ctx, index, document.ID, document)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Elastic) RemoveHashtag(ctx context.Context, index string, hashtagId int) error {
	documents, err := c.GetByHashTagId(ctx, index, hashtagId)
	if err != nil {
		return err
	}
	for idx, document := range documents {
		for idx2, project := range document.Projects {
			remainigHashtags := []model.Hashtag{}
			for _, hashtag := range project.Hashtags {
				if hashtag.ID != hashtagId {
					remainigHashtags = append(remainigHashtags, hashtag)
				}
			}
			documents[idx].Projects[idx2].Hashtags = remainigHashtags
			err = c.Update(ctx, index, document.ID, documents[idx])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// Function to update a document
func (c *Elastic) Update(ctx context.Context, index string, id int, user model.User) error {
	updateResult, err := c.c.Update().
		Index(index).
		Id(fmt.Sprintf("%d", id)).
		Type("_doc").
		Doc(user).
		Do(ctx)
	if err != nil {
		return err
	}
	if updateResult.Result == "updated" {
		return nil
	} else if updateResult.Result == "noop" {
		return fmt.Errorf("No changes made for project with ID %d.\n", id)
	} else {
		return fmt.Errorf("Project update result: %s", updateResult.Result)
	}
}

// Function to delete a document
func (c *Elastic) Delete(ctx context.Context, index string, id int) error {
	_, err := c.c.Delete().
		Index(index).
		Type("_doc").
		Id(fmt.Sprintf("%d", id)).
		Do(ctx)
	return err
}

func (c *Elastic) SearchByUser(ctx context.Context, index string, userID int) (*model.User, error) {
	return c.GetByUserId(ctx, index, userID)
}

func (c *Elastic) SearchByHashtags(ctx context.Context, index string, hashtag string) ([]model.User, error) {
	query := elastic.NewQueryStringQuery(fmt.Sprintf("_source.projects.hashtags.name=%s", hashtag))
	searchService := c.c.Search().Index(index).Query(query)
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, err
	}
	var results []model.User
	for _, hit := range searchResult.Hits.Hits {
		var result model.User
		err := json.Unmarshal(hit.Source, &result)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	return results, nil
}

func (c *Elastic) FuzzySearchProjects(ctx context.Context, index string, query string) ([]model.FuzzyResult, error) {
	qry := elastic.NewBoolQuery().
		Should(
			elastic.NewFuzzyQuery("projects.slug", query).
				Fuzziness(5),
			elastic.NewFuzzyQuery("projects.description", query).
				Fuzziness(5))
	searchService := c.c.Search().Query(qry)
	searchResult, err := searchService.Do(ctx)
	if err != nil {
		return nil, err
	}
	var results []model.FuzzyResult
	for _, hit := range searchResult.Hits.Hits {
		var user model.User
		err := json.Unmarshal(hit.Source, &user)
		if err != nil {
			return nil, err
		}
		hashtags := []string{}
		for _, project := range user.Projects {
			for _, hashtag := range project.Hashtags {
				hashtags = append(hashtags, hashtag.Name)
			}
		}
		results = append(results, model.FuzzyResult{
			Hashtags: hashtags,
			User: model.FuzzyUser{
				ID:        user.ID,
				Name:      user.Name,
				CreatedAt: user.CreatedAt,
			},
		})
	}
	return results, nil
}
