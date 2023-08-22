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

type SearchResult struct {
	Project Project `json:"project"`
	User    User    `json:"user"`
}
