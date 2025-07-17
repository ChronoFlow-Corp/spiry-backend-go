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
	Database   database   `yaml:"database"`
	JWT        jwt        `yaml:"jwt"`
}

type database struct {
	PostgresPassword string `config:"postgresPassword"`
	PostgresHost     string `config:"postgresHost"`
	PostgresPort     string `config:"postgresPort"`
	PostgresUser     string `config:"postgresUser"`
	PostgresDatabase string `config:"postgresDatabase"`
}

type jwt struct {
	RefreshSecret       string        `yaml:"refreshSecret"`
	AccessSecretPublic  string        `yaml:"accessSecretPublic"`
	AccessSecretPrivate string        `yaml:"accessSecretPrivate"`
	AccessExpire        time.Duration `yaml:"accessExpire"`
	RefreshExpire       time.Duration `yaml:"refreshExpire"`
}
type http struct {
	Addr     string        `env:"HTTP_ADDR"       env-default:"127.0.0.1" yaml:"addr"`
	Port     int           `env:"HTTP_PORT"       env-default:"8080"      yaml:"port"`
	Timeout  time.Duration `env:"HTTP_TIMEOUT"    env-default:"5s"        yaml:"timeout"`
	CertFile string        `env:"HTTPS_CERT_FILE"                         yaml:"certFile"`
	KeyFile  string        `env:"HTTPS_KEY_FILE"                          yaml:"keyFile"`
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

	if c.HTTP.CertFile != "" && c.HTTP.KeyFile != "" {
		c.mustSslLoad()
	}
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
