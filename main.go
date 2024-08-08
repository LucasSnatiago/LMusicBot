package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/LucasSnatiago/LMusicBot/config"
	"github.com/bwmarrin/discordgo"
	"github.com/lrstanley/go-ytdlp"
)

var BotConfig *config.BotConfig

func init() {
	// If yt-dlp isn't installed yet, download and cache it for further use.
	ytdlp.MustInstall(context.TODO(), nil)
	BotConfig = config.New()
}

func main() {
	session, err := discordgo.New("Bot " + BotConfig.DiscordToken)
	if err != nil {
		log.Fatal("Bot could not run:", err)
	}

	session.Identify.Intents |= discordgo.IntentsGuildMessages
	session.Identify.Intents |= discordgo.IntentDirectMessages
	session.Identify.Intents |= discordgo.IntentsGuildVoiceStates

	session.AddHandler(LoginMessage)
	session.AddHandler(HandleMessages)

	// Open a websocket connection to Discord and begin listening.
	err = session.Open()
	if err != nil {
		fmt.Println("error opening connection:", err)
		return
	}

	// Graceful shutdown
	fmt.Println("Use CTRL + C to stop the Bot")
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	session.Close()
	fmt.Println("Bot is off")
}

func LoginMessage(s *discordgo.Session, r *discordgo.Ready) {
	log.Printf("Logged in as %s", r.User.String())

	// Set the playing status.
	s.UpdateGameStatus(0, fmt.Sprintf("%shelp", BotConfig.BotPrefix))
}

func HandleMessages(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself and ignore all messages without the bot prefix
	if m.Author.ID == s.State.User.ID && !strings.HasPrefix(m.Content, BotConfig.BotPrefix) {
		return
	}

	// From now on, remove bot prefix
	command := strings.Replace(m.Content, BotConfig.BotPrefix, "", 1)

	switch command {
	case "ping":
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	case "pong":
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	case "help":
		s.ChannelMessageSend(m.ChannelID, "Available commands:\n\nping: bot answers pong!\nhelp: show this message")
	}
}
