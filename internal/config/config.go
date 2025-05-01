package config

import (
	"flag"
	"fmt"
	"net/url"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	DefaultConfigPath = "./config.yaml"
)

type Config struct {
	App        AppConfig      `yaml:"app" env-requried:"true"`
	HTTPServer HTTPConfig     `yaml:"http" env-requried:"true"`
	Database   DatabaseConfig `yaml:"database" env-requried:"true"`
	Logger     LoggerConfig   `yaml:"logger" env-requried:"true"`
}

type AppConfig struct {
	Version string `yaml:"version" env-requried:"true"`

	Balancer    BalancerConfig `yaml:"balancer"`
	RateLimiter RateLimiterConfig `yaml:"rate_limit"`
}

type BalancerConfig struct {
	Servers     []ServerConfig `yaml:"servers"`  // Список бэкендов
	Strategy    string         `yaml:"strategy"` // Стратегия балансировки
	HealthCheck time.Duration  `yaml:"healthCheck" env-default:"10 * time.Second"`
}

type RateLimiterConfig struct {
	DefaultCapacity int           `yaml:"default_capacity" env-default:"100"` // Список бэкендов
	DefaultFillRate time.Duration `yaml:"default_fill_rate" env-default:"1 * time.Second"`
}

type ServerConfig struct {
	ID  int    `yaml:"id"`  // Уникальный ID сервера
	URL string `yaml:"url"` // URL бэкенда
}

type HTTPConfig struct {
	Port              int           `yaml:"port" env-default:"8080"`
	ReadHeaderTimeout time.Duration `yaml:"read_header_timeout" env-requried:"true"`
	ReadTimeout       time.Duration `yaml:"read_timeout" env-requried:"true"`
	WriteTimeout      time.Duration `yaml:"write_timeout" env-requried:"true"`
	IdleTimeout       time.Duration `yaml:"idle_timeout" env-requried:"true"`
}

type DatabaseConfig struct {
	Migration struct {
		Dir    string `yaml:"dir" env-requried:"true"`
		Option string `yaml:"option"`
	} `yaml:"migration" env-requried:"true"`
	Postgres struct {
		URL string `yaml:"url" env-requried:"true"`
	} `yaml:"postgres" env-requried:"true"`
}

type LoggerConfig struct {
	Enable bool   `yaml:"enable" env-default:"true"`
	Level  string `yaml:"level" env-default:"INFO"`  // Avaliable: DEBUG, INFO, WARN, ERROR
	Format string `yaml:"format" env-default:"JSON"` // Avaliable: JSON, TXT
}

// MustLoadCfg.
// fetchConfigPath fetches config path from command line flag or set default variable.
// Priority: flag > default.
// Default value is local config path.
func MustLoadCfg(fetchFlag bool) *Config {
	configPath := DefaultConfigPath
	if fetchFlag {
		p := fetchConfigPath()
		if p != "" {
			configPath = p
		}
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("failed to read config: " + err.Error())
	}

	if err := cfg.Validate(); err != nil {
		panic("invalid config: " + err.Error())
	}

	return &cfg
}

func fetchConfigPath() string {
	var p string
	flag.StringVar(&p, "cfg", "", "path to config")

	flag.Parse()

	return p
}

func (c Config) Validate() error {
	ids := make(map[int]bool)
	for _, s := range c.App.Balancer.Servers {
		if s.ID <= 0 {
			return fmt.Errorf("invalid server ID: %d", s.ID)
		}
		if ids[s.ID] {
			return fmt.Errorf("duplicate server ID: %d", s.ID)
		}
		ids[s.ID] = true

		if _, err := url.Parse(s.URL); err != nil {
			return fmt.Errorf("invalid URL for server %d: %v", s.ID, err)
		}
	}
	return nil
}
