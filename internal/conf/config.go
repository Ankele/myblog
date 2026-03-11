package conf

import (
	"errors"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		HTTP struct {
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"http"`
		GRPC struct {
			Addr    string `yaml:"addr"`
			Timeout string `yaml:"timeout"`
		} `yaml:"grpc"`
	} `yaml:"server"`
	Data struct {
		Database struct {
			Type   string `yaml:"type"`
			Source string `yaml:"source"`
		} `yaml:"database"`
		Redis struct {
			Addr     string `yaml:"addr"`
			Password string `yaml:"password"`
			DB       int    `yaml:"db"`
		} `yaml:"redis"`
	} `yaml:"data"`
	App struct {
		SessionSecret        string `yaml:"session_secret"`
		DefaultAdminUsername string `yaml:"default_admin_username"`
		DefaultAdminPassword string `yaml:"default_admin_password"`
		UploadDir            string `yaml:"upload_dir"`
		FrontendDist         string `yaml:"frontend_dist"`
	} `yaml:"app"`
}

func Load(path string) (*Config, error) {
	cfg := defaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := yaml.NewDecoder(f).Decode(cfg); err != nil {
		return nil, err
	}

	if cfg.Data.Database.Source == "" {
		return nil, errors.New("database source is required")
	}

	if cfg.App.SessionSecret == "" {
		cfg.App.SessionSecret = "change-me-session-secret"
	}
	if cfg.App.DefaultAdminUsername == "" {
		cfg.App.DefaultAdminUsername = "admin"
	}
	if cfg.App.DefaultAdminPassword == "" {
		cfg.App.DefaultAdminPassword = "admin123456"
	}
	if cfg.App.UploadDir == "" {
		cfg.App.UploadDir = "./uploads"
	}
	if cfg.App.FrontendDist == "" {
		cfg.App.FrontendDist = "./web/dist"
	}

	return cfg, nil
}

func (c *Config) HTTPTimeout() time.Duration {
	timeout, err := time.ParseDuration(c.Server.HTTP.Timeout)
	if err != nil {
		return time.Minute
	}
	return timeout
}

func defaultConfig() *Config {
	cfg := &Config{}
	cfg.Server.HTTP.Addr = "0.0.0.0:8000"
	cfg.Server.HTTP.Timeout = "1m"
	cfg.Server.GRPC.Addr = "0.0.0.0:9000"
	cfg.Server.GRPC.Timeout = "1m"
	cfg.App.SessionSecret = "change-me-session-secret"
	cfg.App.DefaultAdminUsername = "admin"
	cfg.App.DefaultAdminPassword = "admin123456"
	cfg.App.UploadDir = "./uploads"
	cfg.App.FrontendDist = "./web/dist"
	return cfg
}
