package config

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/ardanlabs/conf/v2"
	"github.com/joho/godotenv"
)

type App struct {
	conf.Version
	Pg     Pg
	Es     Es
	Server Server
}

type Es struct {
	Host  string `conf:"required"`
	Index string `conf:"default:root"`
}

type Server struct {
	Port int `conf:"default:8080"`
}

type Pg struct {
	Host                         string        `conf:"required"`
	Port                         string        `conf:"required"`
	Username                     string        `conf:"required"`
	Password                     string        `conf:"required"`
	DbName                       string        `conf:"required"`
	DisableTLS                   bool          `conf:"default:true"`
	MaxIdleTimeForConns          time.Duration `conf:"default:30s"`
	MaxLifetimeForConns          time.Duration `conf:"default:1m"`
	MaxIdleConns                 int           `conf:"default:100"`
	MaxOpenConns                 int           `conf:"default:100"`
	ListenerMinReconnectInterval time.Duration `conf:"default:1s"`
	ListenerMaxReconnectInterval time.Duration `conf:"default:2s"`
	ListenerChannel              string        `conf:"required"`
}

func (pg Pg) String() string {
	sslMode := "require"
	if pg.DisableTLS {
		sslMode = "disable"
	}
	q := make(url.Values)
	q.Set("sslmode", sslMode)
	q.Set("timezone", "utc")
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(pg.Username, pg.Password),
		Host:     pg.Host,
		Path:     pg.DbName,
		RawQuery: q.Encode(),
	}
	return u.String()
}

func Load() (App, error) {
	cfg := App{
		Version: conf.Version{
			Build: "develop",
			Desc:  "postgres to elastic pipeline & REST endpoints",
		},
	}
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("error loading .env file", err)
	}
	help, err := conf.Parse("", &cfg)
	if err != nil {
		err = fmt.Errorf("conf.Parse() failed: %w", err)
		if errors.Is(err, conf.ErrHelpWanted) {
			return cfg, fmt.Errorf("help: %v, err: %v", help, err.Error())
		}
		return cfg, err
	}
	return cfg, nil
}
