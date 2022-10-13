package config

import (
	"github.com/fullpipe/bore-server/mail"
	"github.com/kelseyhightower/envconfig"
)

func GetConfig() (Config, error) {
	var config Config
	err := envconfig.Process("bore", &config)
	if err != nil {
		return Config{}, err
	}

	return config, nil
}

type Config struct {
	LiteDB      string `required:"true"`
	BooksDir    string `required:"true"`
	TorrentsDir string `required:"true"`
	Debug       bool

	Server Server
	Mailer mail.MailerConfig
	JWT    JWT
}

type Server struct {
	Port int `default:"8080"`
}

type JWT struct {
	PrivateKey string `required:"true"`
	PublicKey  string `required:"true"`
}
