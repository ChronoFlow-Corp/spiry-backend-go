// Package config provide configuration spiry application.
package config

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	devEnv  = "development"
	prodEnv = "production"
)

// Config struct contains all for start spiry application.
type Config struct {
	Env        string     `env:"ENV"           env-default:"development" yaml:"env"`
	HTTP       http       `yaml:"http"`
	GoogleAuth googleAuth `yaml:"google"`
	Database   database   `env-required:"true" yaml:"database"`
	JWT        jwt        `yaml:"jwt"`
	LLM        llm        `env-required:"true" yaml:"llm"`
}

type database struct {
	PostgresPassword string `env-required:"true"       yaml:"postgres_password"`
	PostgresHost     string `env-required:"true"       yaml:"postgres_host"`
	PostgresPort     string `env-required:"true"       yaml:"postgres_port"`
	PostgresUser     string `env-required:"true"       yaml:"postgres_user"`
	PostgresDatabase string `env-required:"true"       yaml:"postgres_database"`
	PathToMigrations string `yaml:"path_to_migrations"`
}

type jwt struct {
	RefreshSecret       string        `env-required:"true" yaml:"refresh_secret"`
	AccessSecretPublic  string        `env-required:"true" yaml:"access_secret_public"`
	AccessSecretPrivate string        `env-required:"true" yaml:"access_secret_private"`
	AccessExpire        time.Duration `env-default:"3h"    yaml:"access_expire"`
	RefreshExpire       time.Duration `env-default:"24h"   yaml:"refresh_expire"`
}
type http struct {
	Addr        string        `env:"HTTP_ADDR"       env-default:"localhost" yaml:"addr"`
	Port        int           `env:"HTTP_PORT"       env-default:"8080"      yaml:"port"`
	Timeout     time.Duration `env:"HTTP_TIMEOUT"    env-default:"5s"        yaml:"timeout"`
	CertFile    string        `env:"HTTPS_CERT_FILE" yaml:"cert_file"`
	KeyFile     string        `env:"HTTPS_KEY_FILE"  yaml:"key_file"`
	FrontendUrl string        `env:"FRONTEND_URL"    env-required:"true"     yaml:"frontend_url"`
}

type googleAuth struct {
	ClientID     string `env:"GOOGLE_CLIENT_ID"     env-required:"true" yaml:"client_id"`
	ClientSecret string `env:"GOOGLE_CLIENT_SECRET" env-required:"true" yaml:"client_secret"`
	RedirectURI  string `env:"GOOGLE_REDIRECT_URI"  env-required:"true" yaml:"redirect_uri"`
}

type llm struct {
	Key string `env-required:"true" yaml:"key"`
}

func NewConfig() *Config {
	cfg := &Config{}
	cfg.MustLoad()

	return cfg
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

	if c.Env != devEnv && c.Env != prodEnv {
		panic(fmt.Sprintf("Environment variable %s not allowed", c.Env))
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
	defer pbFd.Close() //nolint:errcheck // no matter

	pb, err := io.ReadAll(pbFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read access secret public file: %s", err))
	}

	c.JWT.AccessSecretPublic = string(pb)

	prFd, err := os.Open(c.JWT.AccessSecretPrivate)
	if err != nil {
		panic(fmt.Sprintf("failed to open access secret private file: %s", err))
	}
	defer prFd.Close() //nolint:errcheck // no matter

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
	defer certFd.Close() //nolint:errcheck // no matter

	certBytes, err := io.ReadAll(certFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read ssl cert file: %s: %s", c.HTTP.CertFile, err))
	}

	c.HTTP.CertFile = string(certBytes)

	keyFd, err := os.Open(c.HTTP.KeyFile)
	if err != nil {
		panic(fmt.Sprintf("failed to open ssl key file: %s: %s", c.HTTP.KeyFile, err))
	}
	defer keyFd.Close() //nolint:errcheck // no matter

	keyBytes, err := io.ReadAll(keyFd)
	if err != nil {
		panic(fmt.Sprintf("failed to read ssl key file: %s: %s", c.HTTP.KeyFile, err))
	}

	c.HTTP.KeyFile = string(keyBytes)
}
