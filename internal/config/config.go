package config

import (
	"log"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	TelegramBotToken string
	AdminID          int
	DatabaseURL      string
	HuggingFaceKey   string
	ProxyURL         string
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		log.Println("Config file not found, using environment variables")
	}

	config := &Config{
		TelegramBotToken: viper.GetString("TELEGRAM_BOT_TOKEN"),
		AdminID:          viper.GetInt("ADMIN_ID"),
		DatabaseURL:      viper.GetString("DATABASE_URL"),
		HuggingFaceKey:   os.Getenv("HUGGINGFACE_API_KEY"),
		ProxyURL:         viper.GetString("PROXY_URL"),
	}

	return config, nil
}
