package model

type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	CreatedAt string    `json:"created_at"`
	Projects  []Project `json:"projects"`
}

type Project struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Description string    `json:"description"`
	CreatedAt   string    `json:"created_at"`
	Hashtags    []Hashtag `json:"hashtags"`
}

type Hashtag struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

type UserProject struct {
	ProjectId int    `json:"project_id"`
	UserId    string `json:"user_id"`
}

type ProjectHashtag struct {
	ProjectId int `json:"project_id"`
	HashtagId int `json:"hashtag_id"`
}

type FuzzyResult struct {
	Hashtags []string  `json:"hastags"`
	User     FuzzyUser `json:"user"`
}

type FuzzyUser struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}
