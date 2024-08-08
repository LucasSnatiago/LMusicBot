package config

import (
	"embed"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type BotConfig struct {
	DiscordToken string
	BotOwner     string
	BotPrefix    string
}

//go:embed config_example.txt
var config_file_example embed.FS

func New() *BotConfig {
	var botcfg BotConfig

	err := godotenv.Load()
	if err != nil {
		file, err := os.Create(".env")
		if err != nil {
			fmt.Println("Could not create config file on disk, please check your permissions")
			os.Exit(1)
		}
		defer file.Close()

		config_example, err := config_file_example.ReadFile("config_example.txt")
		if err != nil {
			fmt.Println("Example file not found, something is wrong in compilation!")
			os.Exit(1)
		}

		_, err = file.Write(config_example)
		if err != nil {
			fmt.Println("Could not write to file, please check your permissions")
			os.Exit(1)
		}

		fmt.Println("Please fill in with the necessary info in the .env file, then run the bot again!")
		os.Exit(1)
	}

	botcfg.DiscordToken = os.Getenv("DISCORD_TOKEN")
	botcfg.BotOwner = os.Getenv("BOT_OWNER")
	botcfg.BotPrefix = os.Getenv("BOT_PREFIX")

	return &botcfg
}
