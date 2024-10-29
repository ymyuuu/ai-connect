package config

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var config Config

type Config struct {
	Chatgpt struct {
		Mirror Mirror `json:"mirror"`
	} `json:"chatgpt"`
}

type Mirror struct {
	Address string `json:"address"`
	Tls     struct {
		Enabled bool   `json:"enabled"`
		Key     string `json:"key"`
		Cert    string `json:"cert"`
	} `json:"tls"`
	Tokens map[string]string `json:"tokens"`
}

func Init() {
	err := viper.Unmarshal(&config)
	cobra.CheckErr(err)
}

func ChatGptMirror() Mirror {
	return config.Chatgpt.Mirror
}
