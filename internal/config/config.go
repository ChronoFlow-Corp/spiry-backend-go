// Package config provide configuration spiry application.
package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
	"time"
)

// Config struct contains all for start spiry application.
type Config struct {
	HTTP       http       `yaml:"http"`
	GoogleAuth googleAuth `yaml:"google"`
}

type http struct {
	Addr     string        `env:"HTTP_ADDR"       env-default:"127.0.0.1"                     yaml:"addr"`
	Port     int           `env:"HTTP_PORT"       env-default:"8080"                          yaml:"port"`
	Timeout  time.Duration `env:"HTTP_TIMEOUT"    env-default:"5s"                            yaml:"timeout"`
	CertFile string        `env:"HTTPS_CERT_FILE"                         env-required:"true" yaml:"certFile"`
	KeyFile  string        `env:"HTTPS_KEY_FILE"                          env-required:"true" yaml:"keyFile"`
}

type googleAuth struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID"     env-required:"true" yaml:"clientId"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET" env-required:"true" yaml:"clientSecret"`
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
}
