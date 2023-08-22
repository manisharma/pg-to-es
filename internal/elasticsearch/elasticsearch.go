package elasticsearch

import (
	"pg-to-es/internal/config"

	"github.com/olivere/elastic"
)

type Client struct {
	c *elastic.Client
}

func New(cfg config.Es) (*Client, error) {
	client, err := elastic.NewClient(elastic.SetURL(cfg.Host))
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}
