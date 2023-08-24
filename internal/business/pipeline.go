package business

import (
	"context"
	"encoding/json"
	"log"
	"pg-to-es/internal/contract"
	"pg-to-es/internal/model"
)

type Pipeline struct {
	listener contract.DbListener
	es       contract.Elastic
	index    string
	// TODO: make this pipeline asyc by introducing message brokers, for async processing of delta
}

func NewPipeline(listener contract.DbListener, es contract.Elastic, index string) *Pipeline {
	return &Pipeline{listener, es, index}
}

func (p *Pipeline) Start(ctx context.Context) error {
	deltaStream, err := p.listener.Start(ctx)
	if err != nil {
		return err
	}
	process(ctx, p.es, deltaStream, p.index)
	return err
}

func (p *Pipeline) Stop() {
	p.listener.Stop()
}

func process(ctx context.Context, es contract.Elastic, deltaStream <-chan string, index string) {
	// Process delta
	go func() {
		type payload struct {
			Operation string          `json:"operation"`
			Table     string          `json:"table"`
			Payload   json.RawMessage `json:"payload"`
		}
		type document struct {
			UserID             int    `json:"user_id"`
			UserName           string `json:"user_name"`
			UserCreatedAt      string `json:"user_created_at"`
			ProjectID          int    `json:"project_id"`
			ProjectName        string `json:"project_name"`
			ProjectSlug        string `json:"project_slug"`
			ProjectDescription string `json:"project_description"`
			ProjectCreatedAt   string `json:"project_created_at"`
			HashtagID          int    `json:"hashtag_id"`
			HashtagName        string `json:"hashtag_name"`
			HashtagCreatedAt   string `json:"hashtag_created_at"`
			Operation          string `json:"operation"`
			Table              string `json:"table"`
		}
		for data := range deltaStream {
			log.Println("data", data)
			var d payload
			err := json.Unmarshal([]byte(data), &d)
			if err != nil {
				log.Printf("json.Unmarshal() failed, content: '%s', err: %s", data, err)
				continue
			}

			switch d.Operation {
			case "INSERT", "UPDATE":
				var delta document
				err = json.Unmarshal(d.Payload, &delta)
				if err != nil {
					log.Printf("\njson.Unmarshal(d.Payload, &user), err: %s", err)
					continue
				}
				var esDocx []model.User
				switch {
				case delta.UserID > 0:
					esDoc, _ := es.GetByUserId(ctx, index, delta.UserID)
					if esDoc != nil {
						esDocx = []model.User{
							*esDoc,
						}
					}
				case delta.ProjectID > 0:
					esDocx, _ = es.GetByProjectId(ctx, index, delta.UserID)
				case delta.HashtagID > 0:
					esDocx, _ = es.GetByHashTagId(ctx, index, delta.UserID)
				}
				if esDocx == nil {
					user := &model.User{
						ID:        delta.UserID,
						Name:      delta.UserName,
						CreatedAt: delta.UserCreatedAt,
						Projects:  []model.Project{},
					}
					if delta.ProjectID > 0 {
						project := model.Project{
							ID:          delta.ProjectID,
							Name:        delta.ProjectName,
							Slug:        delta.ProjectSlug,
							Description: delta.ProjectDescription,
							CreatedAt:   delta.ProjectCreatedAt,
							Hashtags:    []model.Hashtag{},
						}
						if delta.HashtagID > 0 {
							project.Hashtags = append(project.Hashtags, model.Hashtag{
								ID:        delta.HashtagID,
								Name:      delta.HashtagName,
								CreatedAt: delta.ProjectCreatedAt,
							})
						}
						user.Projects = append(user.Projects, project)
					}
					err = es.Create(ctx, index, delta.UserID, *user)
					if err != nil {
						log.Printf("\nes.Create() failed, err: %s", err)
						continue
					}
				} else {
					for idx := range esDocx {
						esDocx[idx].Name = delta.ProjectName
						if d.Operation == "UPDATE" {
							for pIdx, project := range esDocx[idx].Projects {
								if project.ID == delta.ProjectID {
									esDocx[idx].Projects[pIdx].Name = delta.ProjectName
									esDocx[idx].Projects[pIdx].Description = delta.ProjectDescription
									esDocx[idx].Projects[pIdx].Slug = delta.ProjectSlug
									for hIdx, hashtag := range esDocx[idx].Projects[pIdx].Hashtags {
										if hashtag.ID == delta.HashtagID {
											esDocx[idx].Projects[pIdx].Hashtags[hIdx].Name = delta.HashtagName
										}
									}
								}
							}
						} else {
							newProject := model.Project{
								ID:          delta.ProjectID,
								Name:        delta.ProjectName,
								Description: delta.ProjectDescription,
								Slug:        delta.ProjectSlug,
								CreatedAt:   delta.ProjectCreatedAt,
							}
							newProject.Hashtags = []model.Hashtag{
								{
									ID:        delta.HashtagID,
									Name:      delta.HashtagName,
									CreatedAt: delta.HashtagCreatedAt,
								},
							}
							if esDocx[idx].Projects == nil {
								esDocx[idx].Projects = []model.Project{}
							}
							esDocx[idx].Projects = append(esDocx[idx].Projects, newProject)
						}
						err = es.Update(ctx, index, delta.UserID, esDocx[idx])
						if err != nil {
							log.Printf("\nes.Update() failed, err: %s", err)
							continue
						}
					}

				}
			case "DELETE":
				switch d.Table {
				case "users":
					var u model.User
					err = json.Unmarshal(d.Payload, &u)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}

					log.Println(u)
					err = es.Delete(ctx, index, u.ID)
					if err != nil {
						log.Printf("\nes.Delete() failed, err: %s", err)
						continue
					}

				case "projects":
					var p model.Project
					err = json.Unmarshal(d.Payload, &p)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}

					err = es.RemoveProject(ctx, index, p.ID)
					if err != nil {
						log.Printf("\nes.RemoveProject() failed, err: %s", err)
						continue
					}

				case "hashtags":
					var h model.Hashtag
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}

					err = es.RemoveHashtag(ctx, index, h.ID)
					if err != nil {
						log.Printf("\nes.RemoveHashtag() failed, err: %s", err)
						continue
					}

				case "project_hashtags":
					var h model.ProjectHashtag
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}

					err = es.RemoveHashtag(ctx, index, h.HashtagId)
					if err != nil {
						log.Printf("\nes.RemoveHashtag() failed, err: %s", err)
						continue
					}
				case "user_projects":
					var h model.UserProject
					err = json.Unmarshal(d.Payload, &h)
					if err != nil {
						log.Printf("json.Unmarshal() failed, err: %v", err)
						continue
					}

					err = es.RemoveProject(ctx, index, h.ProjectId)
					if err != nil {
						log.Printf("\nes.RemoveProject() failed, err: %s", err)
						continue
					}
				}
			}
		}
	}()
}
