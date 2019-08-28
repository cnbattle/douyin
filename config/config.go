package config

import (
	"github.com/spf13/viper"
)

var V *viper.Viper

func init() {
	V = viper.New()
	V.SetConfigName("config")
	V.AddConfigPath("./")
	V.SetConfigType("toml")
	if err := V.ReadInConfig(); err != nil {
		panic(err)
	}
}
