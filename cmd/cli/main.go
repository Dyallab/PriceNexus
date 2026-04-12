package main

import (
	"os"

	"github.com/dyallo/pricenexus/cmd"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

func main() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log := logrus.New()
			log.Warnf("Error loading .env file: %v", err)
		}
	}

	if openRouterKey := os.Getenv("OPENROUTER_API_KEY"); openRouterKey != "" {
		if os.Getenv("OPENAI_API_KEY") == "" {
			os.Setenv("OPENAI_API_KEY", openRouterKey)
		}
	}

	log := logrus.New()
	log.SetLevel(logrus.InfoLevel)

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Error(err)
		os.Exit(1)
	}
}
