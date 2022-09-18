package util

import (
	"github.com/spf13/viper"
	"log"
)

type apiConfig struct {
	Server string `json:"server"`
}

var (
	ApiConfig      apiConfig
	CORSConfig     []string
	RedirectConfig []string
	RefererConfig  map[string]string
)

func LoadConfig() {
	viper.SetConfigFile("config.toml")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalln(err)
	}

	ApiConfig.Server = HandleHost(viper.GetString("api.server"))
	CORSConfig = viper.GetStringSlice("domains.cors")
	RedirectConfig = viper.GetStringSlice("domains.redirect")
	RefererConfig = viper.GetStringMapString("domains.referer")
}
