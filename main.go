package main

import (
	"fmt"
	"os"

	"github.com/Largeb0525/personal-tool/cmd"
	"github.com/Largeb0525/personal-tool/database"
	"github.com/Largeb0525/personal-tool/internal"

	"github.com/spf13/viper"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	db := database.InitDatabase()
	defer db.Close()

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	router := internal.InitRouter()
	if err := router.Run(":" + port); err != nil {
		fmt.Fprintf(os.Stderr, "Error starting server: %v\n", err)
		os.Exit(1)
	}
}
