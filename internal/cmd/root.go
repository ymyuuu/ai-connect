package cmd

import (
	"github.com/dhbin/ai-connect/internal/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
)

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "ai-connect",
		Short: "ai-connect",
	}
)

func init() {
	cobra.OnInitialize(initialConfig)

	chatgptCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is ./config.json)")
	chatgptCmd.PersistentFlags().BoolVarP(&mirror, "mirror", "", false, "chatgpt镜像")
	rootCmd.AddCommand(chatgptCmd)
}

func initialConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(".")
		viper.AddConfigPath(home + "/.ai-connect")
		viper.SetConfigType("json")
		viper.SetConfigName("config")

		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			cobra.CheckErr(err.Error())
		}
		config.Init()
	}
}

func Execute() {
	err := rootCmd.Execute()
	cobra.CheckErr(err)
}
