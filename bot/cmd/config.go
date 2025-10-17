package main

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type TelegramConfig struct {
	BotToken   string `mapstructure:"BOT_TOKEN"`
	BotApiUrl  string `mapstructure:"BOT_API_URL"`
	WebhookUrl string `mapstructure:"WEBHOOK_URL"`
}

type DbConfig struct {
	Host     string `mapstructure:"HOST"`
	Port     int    `mapstructure:"PORT"`
	User     string `mapstructure:"USER"`
	Password string `mapstructure:"PASSWORD"`
	Name     string `mapstructure:"NAME"`
}

type Config struct {
	MaxWorkers int            `mapstructure:"MAX_WORKERS"`
	Telegram   TelegramConfig `mapstructure:"TELEGRAM"`
	Db         DbConfig       `mapstructure:"DB"`
}

func LoadConfig(path string) (Config, error) {
	var cfg Config
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(path)
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return cfg, fmt.Errorf("couldn't read config: %w", err)
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return cfg, fmt.Errorf("couldn't unmarshal the config: %w", err)
	}
	return cfg, nil
}

func ValidateConfig(cfg *Config) bool {
	// ensure all the fields are non empty
	if cfg.Telegram.BotApiUrl == "" || cfg.Telegram.BotToken == "" || cfg.Telegram.WebhookUrl == "" || cfg.Db.Host == "" || cfg.Db.Port == 0 || cfg.Db.User == "" || cfg.Db.Name == "" || cfg.Db.Password == "" {
		return false
	}
	return true
}
