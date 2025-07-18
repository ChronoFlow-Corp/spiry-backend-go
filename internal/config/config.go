// Package config provide configuration spiry application.
package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config struct contains all for start spiry application.
type Config struct {
	HTTP       http       `yaml:"http"`
	GoogleAuth googleAuth `yaml:"google"`
	Database   database   `yaml:"database" env-required:"true"`
	JWT        jwt        `yaml:"jwt"`
}

type database struct {
	PostgresPassword string `yaml:"postgresPassword" env-required:"true"`
	PostgresHost     string `yaml:"postgresHost" env-required:"true"`
	PostgresPort     string `yaml:"postgresPort" env-required:"true"`
	PostgresUser     string `yaml:"postgresUser" env-required:"true"`
	PostgresDatabase string `yaml:"postgresDatabase" env-required:"true"`
}

type jwt struct {
	RefreshSecret       string        `yaml:"refreshSecret" env-required:"true"`
	AccessSecretPublic  string        `yaml:"accessSecretPublic" env-required:"true"`
	AccessSecretPrivate string        `yaml:"accessSecretPrivate" env-required:"true"`
	AccessExpire        time.Duration `yaml:"accessExpire" env-default:"3h"`
	RefreshExpire       time.Duration `yaml:"refreshExpire" env-default:"24h"`
}
type http struct {
	Addr     string        `env:"HTTP_ADDR"       env-default:"localhost" yaml:"addr"`
	Port     int           `env:"HTTP_PORT"       env-default:"8080"      yaml:"port"`
	Timeout  time.Duration `env:"HTTP_TIMEOUT"    env-default:"5s"        yaml:"timeout"`
	CertFile string        `env:"HTTPS_CERT_FILE"                         yaml:"certFile"`
	KeyFile  string        `env:"HTTPS_KEY_FILE"                          yaml:"keyFile"`
	FrontendURL string `env:"FRONTEND_URL" yaml:"frontendURL" env-required:"true"`
}

type googleAuth struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID"     env-required:"true" yaml:"clientId"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET" env-required:"true" yaml:"clientSecret"`
	RedirectURI  string `env:"GOOGLE_REDIRECT_URI" env-required:"true" yaml:"redirectURI"`
}

// MustLoad modify config struct if you have error it panics.
func (c *Config) MustLoad() {
	p := os.Getenv("CONFIG_PATH")
	if p == "" {
		panic("CONFIG_PATH environment variable not set")
	}

	err := cleanenv.ReadConfig(p, c)
	if err != nil {
		panic("failed to read config: " + err.Error())
	}

	c.mustJwtLoad()

	if c.HTTP.CertFile != "" && c.HTTP.KeyFile != "" {
		c.mustSslLoad()
	}
}

func (c *Config) mustJwtLoad() {
	pbFd, err := os.Open(c.JWT.AccessSecretPublic)
	if err != nil {
		panic(fmt.Sprintf("failed to open access secret public file: %s", err))
	}
	defer pbFd.Close()

	pb, err := io.ReadAll(pbFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read access secret public file: %s", err))
	}

	c.JWT.AccessSecretPublic = string(pb)

	prFd, err := os.Open(c.JWT.AccessSecretPrivate)
	if err != nil {
		panic(fmt.Sprintf("failed to open access secret private file: %s", err))
	}
	defer prFd.Close()

	pr, err := io.ReadAll(prFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read access secret private file: %s", err))
	}

	c.JWT.AccessSecretPrivate = string(pr)
}

func (c *Config) mustSslLoad() {
	certFd, err := os.Open(c.HTTP.CertFile)
	if err != nil {
		panic(fmt.Sprintf("failed to open ssl cert file: %s: %s", c.HTTP.CertFile, err))
	}
	defer certFd.Close()

	certBytes, err := io.ReadAll(certFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read ssl cert file: %s: %s", c.HTTP.CertFile, err))
	}
	c.HTTP.CertFile = string(certBytes)

	keyFd, err := os.Open(c.HTTP.KeyFile)
	if err != nil {
		panic(fmt.Sprintf("failed to open ssl key file: %s: %s", c.HTTP.KeyFile, err))
	}
	defer keyFd.Close()

	keyBytes, err := io.ReadAll(keyFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read ssl key file: %s: %s", c.HTTP.KeyFile, err))
	}
	c.HTTP.KeyFile = string(keyBytes)
}
