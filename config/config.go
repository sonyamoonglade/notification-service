package config

import (
	"errors"
	"os"

	"github.com/spf13/viper"
)

const (
	DatabaseURL = "DB_URL"
	BotToken    = "BOT_TOKEN"
	Env         = "ENV"
)

type AppConfig struct {
	DatabaseURL string
	BotToken    string
	AppPort     string
	Env         string
}

func GetAppConfig() (AppConfig, error) {

	v, err := readConfig()
	if err != nil {
		return AppConfig{}, err
	}

	dbURL, ok := os.LookupEnv(DatabaseURL)
	if ok != true {
		return AppConfig{}, errors.New("missing database url")
	}

	botToken, ok := os.LookupEnv(BotToken)
	if ok != true {
		return AppConfig{}, errors.New("missing bot token")
	}

	appPort := v.GetString("app.port")
	if appPort == "" {
		return AppConfig{}, errors.New("missing app.port")
	}
	env, ok := os.LookupEnv(Env)
	if ok != true {
		return AppConfig{}, errors.New("missing env variable")
	}

	return AppConfig{
		DatabaseURL: dbURL,
		BotToken:    botToken,
		AppPort:     appPort,
		Env:         env,
	}, nil
}

func readConfig() (*viper.Viper, error) {

	env, ok := os.LookupEnv(Env)
	if ok != true {
		return nil, errors.New("missing env variable")
	}

	name := "config"

	if env == "production" {
		name = "prod." + name
	} else {
		name = "dev." + name
	}

	viper.AddConfigPath(".")
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	return viper.GetViper(), nil

}
