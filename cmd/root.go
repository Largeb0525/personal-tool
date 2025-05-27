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
		fmt.Println("CLI Ready")
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
		viper.SetConfigName("config")
		viper.AddConfigPath("./config")
		viper.SetConfigType("toml")
		viper.AutomaticEnv()

		_ = viper.BindEnv("database.host", "DATABASE_HOST")
		_ = viper.BindEnv("database.port", "DATABASE_PORT")
		_ = viper.BindEnv("database.user", "DATABASE_USER")
		_ = viper.BindEnv("database.password", "DATABASE_PASSWORD")
		_ = viper.BindEnv("database.dbname", "DATABASE_DBNAME")

		if err := viper.ReadInConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config file: %v\n", err)
			os.Exit(1)
		}
	}
}
