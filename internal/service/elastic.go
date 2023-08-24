package service

import (
	"context"
	"encoding/json"
	"fmt"
	"pg-to-es/internal/config"
	"pg-to-es/internal/model"

	"github.com/olivere/elastic"
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
	userAgg := elastic.NewNestedAggregation().Path("user")
	projectAgg := elastic.NewNestedAggregation().Path("user.projects")
	hashtagAgg := elastic.NewNestedAggregation().Path("user.projects.hashtags")
	hashtagQuery := elastic.NewNestedQuery("user.projects.hashtags",
		elastic.NewTermQuery("user.projects.id", projectId))

	searchResult, err := c.c.Search().
		Index(index).
		Query(hashtagQuery).
		Aggregation("user", userAgg).
		Aggregation("projects", projectAgg).
		Aggregation("hashtags", hashtagAgg).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	// Process and return the search results
	var results []model.User
	for _, hit := range searchResult.Hits.Hits {
		var result model.User
		var m map[string]interface{}
		err := json.Unmarshal(*hit.Source, &m)
		if err != nil {
			return nil, err
		}
		fmt.Println(m)
		results = append(results, result)
	}
	return results, nil
}

func (c *Elastic) GetByHashTagId(ctx context.Context, index string, hashTagId int) ([]model.User, error) {
	userAgg := elastic.NewNestedAggregation().Path("user")
	projectAgg := elastic.NewNestedAggregation().Path("user.projects")
	hashtagAgg := elastic.NewNestedAggregation().Path("user.projects.hashtags")
	hashtagQuery := elastic.NewNestedQuery("user.projects.hashtags",
		elastic.NewTermQuery("user.projects.hashtags.id", hashTagId))

	searchResult, err := c.c.Search().
		Index(index).
		Query(hashtagQuery).
		Aggregation("user", userAgg).
		Aggregation("projects", projectAgg).
		Aggregation("hashtags", hashtagAgg).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	// Process and return the search results
	var results []model.User
	for _, hit := range searchResult.Hits.Hits {
		var result model.User
		var m map[string]interface{}
		err := json.Unmarshal(*hit.Source, &m)
		if err != nil {
			return nil, err
		}
		fmt.Println(m)
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
		err = json.Unmarshal(*doc.Source, &u)
		if err != nil {
			return nil, err
		}
		return &u, nil
	}
	return nil, fmt.Errorf("not found")
}

func (c *Elastic) RemoveProject(ctx context.Context, index string, projectId int) error {
	// Create a script to remove the project from the array
	script := elastic.NewScript("ctx._source.user.projects.removeIf(p -> p.id == params.projectId)")
	script.Params(map[string]interface{}{
		"projectId": projectId,
	})

	// Create an update by query request
	updateByQueryRequest := c.c.UpdateByQuery().
		Index(index).                                               // Replace with your actual index name
		Query(elastic.NewTermQuery("user.projects.id", projectId)). // Match documents with the specified project ID
		Script(script)                                              // Use the script to remove the project from the array

	// Execute the update by query request
	updateResult, err := updateByQueryRequest.Do(ctx)
	if err != nil {
		return err
	}

	// Check the update result
	if updateResult.Updated == 0 {
		return fmt.Errorf("No documents updated.")
	}
	return nil
}

func (c *Elastic) RemoveHashtag(ctx context.Context, index string, hashtagId int) error {
	// Create a script to remove the project from the array
	script := elastic.NewScript("ctx._source.user.projects.hashtags.removeIf(p -> p.id == params.hashtagId)")
	script.Params(map[string]interface{}{
		"hashtagId": hashtagId,
	})

	// Create an update by query request
	updateByQueryRequest := c.c.UpdateByQuery().
		Index(index).                                                        // Replace with your actual index name
		Query(elastic.NewTermQuery("user.projects.hashtags.id", hashtagId)). // Match documents with the specified hashtag ID
		Script(script)                                                       // Use the script to remove the project from the array

	// Execute the update by query request
	updateResult, err := updateByQueryRequest.Do(ctx)
	if err != nil {
		return err
	}

	// Check the update result
	if updateResult.Updated == 0 {
		return fmt.Errorf("No documents updated.")
	}
	return nil
}

// Function to update a document
func (c *Elastic) Update(ctx context.Context, index string, id int, user model.User) error {
	// Create an update request for the specific document by its project ID
	updateResult, err := c.c.Update().
		Index(index).              // Replace with your actual index name
		Id(fmt.Sprintf("%d", id)). // Convert projectID to string and use it as the document ID
		Type("_doc").              // The document type (usually "_doc" for Elasticsearch 7.x)
		Doc(user).                 // Use the update script to modify the document
		Do(ctx)
	if err != nil {
		return err
	}

	// Check the update result
	if updateResult.Result == "updated" {
		return fmt.Errorf("Project with ID %d updated successfully.\n", id)
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

func (c *Elastic) SearchByUser(ctx context.Context, index string, userID int) ([]model.SearchResult, error) {
	userQuery := elastic.NewTermQuery("user.id", userID)
	userAgg := elastic.NewNestedAggregation().Path("user")
	projectAgg := elastic.NewNestedAggregation().Path("user.projects")
	hashtagAgg := elastic.NewNestedAggregation().Path("user.projects.hashtags")

	searchResult, err := c.c.Search().
		Index(index).
		Query(userQuery).
		Aggregation("user", userAgg).
		Aggregation("projects", projectAgg).
		Aggregation("hashtags", hashtagAgg).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	// Process and return the search results
	var results []model.SearchResult
	for _, hit := range searchResult.Hits.Hits {
		var result model.SearchResult
		var m map[string]interface{}
		err := json.Unmarshal(*hit.Source, &m)
		if err != nil {
			return nil, err
		}
		fmt.Println(m)
		results = append(results, result)
	}
	return results, nil
}

func (c *Elastic) SearchByHashtags(ctx context.Context, index string, hashtag string) ([]model.SearchResult, error) {
	userAgg := elastic.NewNestedAggregation().Path("user")
	projectAgg := elastic.NewNestedAggregation().Path("user.projects")
	hashtagAgg := elastic.NewNestedAggregation().Path("user.projects.hashtags")
	hashtagQuery := elastic.NewNestedQuery("user.projects.hashtags",
		elastic.NewTermQuery("user.projects.hashtags.name", hashtag))

	searchResult, err := c.c.Search().
		Index(index).
		Query(hashtagQuery).
		Aggregation("user", userAgg).
		Aggregation("projects", projectAgg).
		Aggregation("hashtags", hashtagAgg).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}
	// Process and return the search results
	var results []model.SearchResult
	for _, hit := range searchResult.Hits.Hits {
		var result model.SearchResult
		var m map[string]interface{}
		err := json.Unmarshal(*hit.Source, &m)
		if err != nil {
			return nil, err
		}
		fmt.Println(m)
		results = append(results, result)
	}
	return results, nil
}

func (c *Elastic) FuzzySearchProjects(ctx context.Context, index string, query string) ([]model.SearchResult, error) {
	userAgg := elastic.NewNestedAggregation().Path("user")
	projectAgg := elastic.NewNestedAggregation().Path("user.projects")
	hashtagAgg := elastic.NewNestedAggregation().Path("user.projects.hashtags")
	fuzzyQuery := elastic.NewBoolQuery().
		Should(
			elastic.NewFuzzyQuery("user.projects.slug", "desired_search_term").
				Boost(2.0),
			elastic.NewFuzzyQuery("user.projects.description", "desired_search_term").
				Boost(1.0))

	searchResult, err := c.c.Search().
		Index(index).
		Query(fuzzyQuery).
		Aggregation("user", userAgg).
		Aggregation("projects", projectAgg).
		Aggregation("hashtags", hashtagAgg).
		Size(10).
		Do(ctx)
	if err != nil {
		return nil, err
	}

	// Process and return the search results
	var results []model.SearchResult
	for _, hit := range searchResult.Hits.Hits {
		var result model.SearchResult
		var m map[string]interface{}
		err := json.Unmarshal(*hit.Source, &m)
		if err != nil {
			return nil, err
		}
		fmt.Println(m)
		results = append(results, result)
	}
	return results, nil
}
