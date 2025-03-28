package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCmd = &cobra.Command{
	Use:   "personal-tool",
	Short: "personal-tool is a tool for personal usage",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello from Cobra CLI!")
	},
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().String("config", "", "config file (default is ./config.toml)")
}

func initConfig() {
	configFile, _ := RootCmd.Flags().GetString("config")
	if configFile != "" {
		viper.SetConfigFile(configFile)
	} else {
		viper.SetConfigName("local")
		viper.AddConfigPath("./config")
		viper.SetConfigType("toml")

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(1)
		}
	}
}
